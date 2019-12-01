package cxrpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxnoise"
)

// OpencxClient is an interface defining the methods a client should implement.
// This could be changed, but right now this is what a client is, and this is what benchclient supports.
// This abstraction only allows us to use either authenticated or unauthenticated clients.
type OpencxClient interface {
	// Call calls the service method with a name, arguments, and reply
	Call(string, interface{}, interface{}) error
	// SetupConnection sets up a connection with the server
	SetupConnection(string, uint16) error
}

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	Conn *rpc.Client
}

// OpencxNoiseClient is an authenticated RPC Client for the opencx Server
type OpencxNoiseClient struct {
	Conn *rpc.Client
	key  *koblitz.PrivateKey
}

// Call calls the servicemethod with name string, args args, and reply reply
func (cl *OpencxRPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.Conn.Call(serviceMethod, args, reply)
}

// SetupConnection creates a new RPC client
func (cl *OpencxRPCClient) SetupConnection(server string, port uint16) (err error) {

	serverAddr := net.JoinHostPort(server, fmt.Sprintf("%d", port))

	if cl.Conn, err = rpc.Dial("tcp", serverAddr); err != nil {
		return
	}

	return
}

// Call calls the servicemethod with name string, args args, and reply reply
func (cl *OpencxNoiseClient) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {

	// Create new client because the server is a rpc newserver
	// we don't need to do anything fancy here because the cl.Conn
	// already uses the noise protocol.
	if err = cl.Conn.Call(serviceMethod, args, reply); err != nil {
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

	serverAddr := net.JoinHostPort(server, fmt.Sprintf("%d", port))

	// Dial a connection to the server
	var clientConn *cxnoise.Conn
	if clientConn, err = cxnoise.Dial(cl.key, serverAddr, []byte("opencx"), net.Dial); err != nil {
		return
	}

	cl.Conn = rpc.NewClient(clientConn)

	return
}
