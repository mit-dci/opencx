package main

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

// Register registers the user for an account with a username and password
func(cl *openCxClient) Register(args []string) error {

	registerArgs := new(cxrpc.RegisterArgs)
	registerReply := new(cxrpc.RegisterReply)

	registerArgs.Username = args[0]
	registerArgs.Password = args[1]

	err := cl.Call("OpencxRPC.Register", registerArgs, registerReply)
	if err != nil {
		return fmt.Errorf("Error calling 'Register' service method:\n%s", err)
	}

	fmt.Printf("Reply: \n%s\n", registerReply)

	return nil
}

func(cl *openCxClient) Login(args []string) error {
	// TODO
	return nil
}
