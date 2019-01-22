package cxrpc

import (
	"net/rpc"
	"encoding/hex"
)

var (
	defaultServer = "localhost"
	defaultPort = 1234
)

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	token  []byte
}

// Register registers the user for an account with a username and password
func(cl *OpencxRPCClient) Register(username string, password string) error {
	var err error

	args := &RegisterArgs{username, password}
	var token []byte

	client, err := rpc.Dial("tcp", defaultServer + ":" + string(defaultPort))
	client.Call("OpencxRPC.Register", args, token)

	res, err := hex.DecodeString("sampleToken")
	if err != nil {
		return err
	}
	cl.token = res
	return nil
}
