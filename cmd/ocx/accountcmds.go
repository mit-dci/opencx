package main

import (
	"github.com/mit-dci/opencx/logging"
)

// Register registers the user for an account with a username and password
func (cl *openCxClient) Register(args []string) (err error) {
	username := args[0]

	// if there is ever a reply for register uncomment this and replace the _
	// var registerReply *cxrpc.RegisterReply
	if _, err = cl.RPCClient.Register(username); err != nil {
		return
	}

	logging.Infof("Successfully registered\n")
	return
}
