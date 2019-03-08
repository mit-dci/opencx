package main

import (
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// Register registers the user for an account with a username and password
func (cl *openCxClient) Register(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}
	var regStringReply *cxrpc.GetRegistrationStringReply
	if regStringReply, err = cl.RPCClient.GetRegistrationString(); err != nil {
		return
	}

	var sig []byte
	if sig, err = cl.SignBytes([]byte(regStringReply.RegistrationString)); err != nil {
		return
	}

	// if there is ever a reply for register uncomment this and replace the _
	// var registerReply *cxrpc.RegisterReply
	if _, err = cl.RPCClient.Register(sig); err != nil {
		return
	}

	logging.Infof("Successfully registered\n")
	return
}
