package cxrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/logging"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxnoise"
)

// OpencxClient is an interface defining the methods a client should implement.
// This could be changed, but right now this is what a client is, and this is what benchclient supports.
// This abstraction only allows us to use either authenticated or unauthenticated clients.
type OpencxClient interface {
	Call(string, interface{}, interface{}) error
	SetupConnection(string, uint16) error
}

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	Conn *rpc.Client
}

// OpencxNoiseClient is an authenticated RPC Client for the opencx Server
type OpencxNoiseClient struct {
	Conn               *cxnoise.Conn
	key                *koblitz.PrivateKey
	requestNonce       uint64
	requestNonceMtx    sync.Mutex
	responseChannelMtx sync.Mutex
	responseChannels   map[uint64]chan lnutil.RemoteControlRpcResponseMsg
	conMtx             sync.Mutex
}

// Call calls the servicemethod with name string, args args, and reply reply
func (cl *OpencxRPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.Conn.Call(serviceMethod, args, reply)
}

// SetupConnection creates a new RPC client
func (cl *OpencxRPCClient) SetupConnection(server string, port uint16) (err error) {

	if cl.Conn, err = rpc.Dial("tcp", server+":"+fmt.Sprintf("%d", port)); err != nil {
		return
	}

	return
}

// Call calls the servicemethod with name string, args args, and reply reply
func (cl *OpencxNoiseClient) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {

	// Generate a local unique nonce using the mutex
	cl.requestNonceMtx.Lock()
	cl.requestNonce++
	nonce := cl.requestNonce
	cl.requestNonceMtx.Unlock()

	// Create the channel to receive a reply on
	cl.responseChannelMtx.Lock()
	cl.responseChannels[nonce] = make(chan lnutil.RemoteControlRpcResponseMsg)
	cl.responseChannelMtx.Unlock()

	// send the message in a goroutine
	go func() {
		msg := new(lnutil.RemoteControlRpcRequestMsg)
		msg.Args, err = json.Marshal(args)
		msg.Idx = nonce
		msg.Method = serviceMethod

		if err != nil {
			logging.Fatal(err)
		}

		rawMsg := msg.Bytes()
		cl.conMtx.Lock()
		n, err := cl.Conn.Write(rawMsg)
		cl.conMtx.Unlock()
		if err != nil {
			logging.Fatal(err)
		}

		if n < len(rawMsg) {
			logging.Fatal(fmt.Errorf("Did not write entire message to peer"))
		}
	}()

	// If reply is nil the caller apparently doesn't care about the results. So we shouldn't wait for it
	if reply != nil {
		// If not nil, await the reply from the responseChannel for the nonce we sent out.
		// the server will include the same nonce in its reply.
		select {
		case receivedReply := <-cl.responseChannels[nonce]:
			{
				if receivedReply.Error {
					err = errors.New(string(receivedReply.Result))
					return
				}

				err = json.Unmarshal(receivedReply.Result, &reply)
				return
			}
		case <-time.After(time.Second * 10):
			// If no reply is received within 10 seconds, we time out the request.
			// TODO: We could make this configurable in the call
			err = errors.New("RPC call timed out")
			return
		}
	}

	return
}

// SetKey sets the private key for the noise client.
func (cl *OpencxNoiseClient) SetKey(privkey *koblitz.PrivateKey) (err error) {
	if privkey == nil {
		err = fmt.Errorf("Cannot set nil key")
		return
	}
	cl.key = privkey
	return
}

// SetupConnection creates a new RPC client
func (cl *OpencxNoiseClient) SetupConnection(server string, port uint16) (err error) {

	if cl.key == nil {
		err = fmt.Errorf("Please set the key for the noise client to create a connection")
		return
	}

	// Create a map of chan objects to receive returned responses on. These channels
	// are sent to from the ReceiveLoop, and awaited in the Call method.
	cl.responseChannels = make(map[uint64]chan lnutil.RemoteControlRpcResponseMsg)

	// Dial a connection to the server
	if cl.Conn, err = cxnoise.Dial(cl.key, server, []byte("opencx"), net.Dial); err != nil {
		return
	}

	// Start the receive loop for reply messages
	go cl.ReceiveLoop()
	return
}

// ReceiveLoop reads messages from the CXNOISE connection and check if they are
// RPC responses
func (cl *OpencxNoiseClient) ReceiveLoop() {
	for {
		msg := make([]byte, 1<<24)
		n, err := cl.Conn.Read(msg)
		if err != nil {
			logging.Warnf("Error reading message from CXNOISE: %s\n", err.Error())
			cl.Conn.Close()
			return
		}
		msg = msg[:n]
		// We only care about RPC responses (for now)
		if msg[0] == lnutil.MSGID_REMOTE_RPCRESPONSE {
			// Parse the received message
			response, err := lnutil.NewRemoteControlRpcResponseMsgFromBytes(msg, 0)
			if err != nil {
				logging.Warnf("Error while receiving RPC response: %s\n", err.Error())
				cl.Conn.Close()
				return
			}

			// Find the response channel to send the reply to
			responseChan, ok := cl.responseChannels[response.Idx]
			if ok {
				// Send the response, but don't depend on someone
				// listening. The caller decides if he's interested in the
				// reply and therefore, it could have not blocked and just
				// ignore the return value.
				select {
				case responseChan <- response:
				default:
				}

				// Clean up the channel to preserve memory. It's only used once.
				cl.responseChannelMtx.Lock()
				delete(cl.responseChannels, response.Idx)
				cl.responseChannelMtx.Unlock()

			} else {
				logging.Errorf("Could not find response channel for index %d\n", response.Idx)
			}
		}
	}
}
