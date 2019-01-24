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

	if registerReply.Token == nil {
		return fmt.Errorf("Username already exists")
	}

	cl.Printf("Successfully registered\n")
	cl.Username = registerArgs.Username
	cl.Token = registerReply.Token
	return nil
}

func(cl *openCxClient) Login(args []string) error {

	loginArgs := new(cxrpc.LoginArgs)
	loginReply := new(cxrpc.LoginReply)

	loginArgs.Username = args[0]
	loginArgs.Password = args[1]

	err := cl.Call("OpencxRPC.Login", loginArgs, loginReply)
	if err != nil {
		return fmt.Errorf("Error calling 'Login' service method:\n%s", err)
	}

	if loginReply.Token == nil {
		return fmt.Errorf("Credentials incorrect")
	}

	cl.Printf("Successfully logged in\n")
	cl.Username = loginArgs.Username
	cl.Token = loginReply.Token
	return nil
}
