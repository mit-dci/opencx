package cxrpc

import (
	"net/rpc"
	"encoding/hex"
)

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	Server string
	Port   int
	token  []byte
}

// Register registers the user for an account with a username and password
func(cl *OpencxRPCClient) Register(username string, password string) error {
	var err error

	args := &RegisterArgs{username, password}
	var token []byte

	client, err := rpc.Dial("tcp", cl.Server + ":" + string(cl.Port))
	if err != nil {
		return err
	}
	err = client.Call("OpencxRPC.Register", args, token)
	if err != nil {
		return err
	}

	res, err := hex.DecodeString("sampleToken")
	if err != nil {
		return err
	}
	cl.token = res
	return nil
}
