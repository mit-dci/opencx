package main

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

type authFunction func(args []string) error

func (cl *openCxClient) AuthCommand(args []string, fn authFunction) error {
	authArgs := new(cxrpc.AuthArgs)
	authReply := new(cxrpc.AuthReply)

	authArgs.Username = cl.Username
	authArgs.Token = cl.Token

	err := cl.Call("OpencxRPC.DoAuthenticatedThing", authArgs, authReply)
	if err != nil {
		return fmt.Errorf("Error when authenticating client: \n%s", err)
	}

	if !authReply.Success {
		cl.Printf("Authentication failed for function: %v\n", fn)
		return nil
	}

	return fn(args)
}

func (cl *openCxClient) uselessFunction(args []string) error {
	cl.Printf("Useless function succeeded\n")
	cl.Printf("%s\n", args)
	return nil
}
