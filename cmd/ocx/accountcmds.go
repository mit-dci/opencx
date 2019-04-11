package main

import (
	"fmt"

	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

var registerCommand = &Command{
	Format: fmt.Sprintf("%s\n", lnutil.Red("register")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Register the public key associated with your private key as an identity on the exchange.",
		"You will use your private key to sign commands that require authorization.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Register yourself on the exchange"),
}

// Register registers the user for an account with a username and password
func (cl *ocxClient) Register(args []string) (err error) {
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
