package cxrpc

import (
	"fmt"
	"net/rpc"
	"encoding/hex"
)

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	conn   *rpc.Client
	token  []byte
}

// Register registers the user for an account with a username and password
func(cl *OpencxRPCClient) Register(args []string) error {
	var err error

	// registercall := client.Go("OpencxRPC.Register", args, token, nil)
	// replycall := <- registercall.Done
	// fmt.Println(token)
	// err = replycall.Error
	// if err != nil {
	//	return err
	// }

	if err != nil {
		return err
	}

	s := fmt.Sprintf("%x", "sampleToken")
	res, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	cl.token = res
	return nil
}

// NewOpencxRPCClient creates a new RPC client
func NewOpencxRPCClient(server string, port int) (*OpencxRPCClient, error) {
	var err error

	cl := new(OpencxRPCClient)

	cl.conn, err = rpc.Dial("tcp", server + ":" + fmt.Sprintf("%d",port))
	println(server + ":" + fmt.Sprintf("%d", port))
	if err != nil {
		return nil, err
	}

	return cl, nil
}
