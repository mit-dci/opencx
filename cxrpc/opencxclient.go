package cxrpc

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"

	"github.com/mit-dci/lit/lnutil"

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

	// Create new client because the server is a rpc newserver
	// we don't need to do anything fancy here because the cl.Conn
	// already uses the noise protocol.
	noiseRPCClient := rpc.NewClient(cl.Conn)

	if err = noiseRPCClient.Call(serviceMethod, args, reply); err != nil {
		return
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

// SetupConnection creates a new RPC Noise client
func (cl *OpencxNoiseClient) SetupConnection(server string, port uint16) (err error) {

	if cl.key == nil {
		err = fmt.Errorf("Please set the key for the noise client to create a connection")
		return
	}

	serverAddr := server + ":" + fmt.Sprintf("%d", port)
	// Create a map of chan objects to receive returned responses on. These channels
	// are sent to from the ReceiveLoop, and awaited in the Call method.
	cl.responseChannels = make(map[uint64]chan lnutil.RemoteControlRpcResponseMsg)

	// Dial a connection to the server
	if cl.Conn, err = cxnoise.Dial(cl.key, serverAddr, []byte("opencx"), net.Dial); err != nil {
		return
	}

	return
}
