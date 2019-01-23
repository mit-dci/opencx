package cxrpc

import (
	"fmt"
	"net/rpc"
)

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	Conn   *rpc.Client
	token  []byte
}

// Call calls the servicemethod with name stirng, args args, and reply reply
func(cl *OpencxRPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.Conn.Call(serviceMethod, args, reply)
}

// NewOpencxRPCClient creates a new RPC client
func NewOpencxRPCClient(server string, port int) (*OpencxRPCClient, error) {
	var err error

	cl := new(OpencxRPCClient)

	cl.Conn, err = rpc.Dial("tcp", server + ":" + fmt.Sprintf("%d",port))
	println(server + ":" + fmt.Sprintf("%d", port))
	if err != nil {
		return nil, err
	}
	println("Dial succeeded")

	return cl, nil
}
