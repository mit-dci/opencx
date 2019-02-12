package main

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

// Register registers the user for an account with a username and password
func (cl *openCxClient) Register(args []string) error {

	registerArgs := new(cxrpc.RegisterArgs)
	registerReply := new(cxrpc.RegisterReply)

	registerArgs.Username = args[0]

	err := cl.Call("OpencxRPC.Register", registerArgs, registerReply)
	if err != nil {
		return fmt.Errorf("Error calling 'Register' service method:\n%s", err)
	}

	cl.Printf("Successfully registered\n")
	return nil
}
