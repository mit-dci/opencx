package cxrpc

import (
	"fmt"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
)

// RegisterArgs holds the args for register
type RegisterArgs struct {
	Signature []byte
}

// RegisterReply holds the data for the register reply
type RegisterReply struct {
	// empty
}

// Register registers a pubkey into the db, verifies that the action was signed by that pubkey. A valid signature for the string "register" is considered a valid registration.
func (cl *OpencxRPC) Register(args RegisterArgs, reply *RegisterReply) (err error) {

	var pubkey *koblitz.PublicKey
	if pubkey, err = cl.Server.RegistrationStringVerify(args.Signature); err != nil {
		err = fmt.Errorf("Error verifying registration string for register RPC command: %s", err)
		return
	}

	if err = cl.Server.RegisterUser(pubkey); err != nil {
		err = fmt.Errorf("Error registering user for register RPC command: %s", err)
		return
	}

	logging.Infof("Registering user with pubkey %x\n", pubkey.SerializeCompressed())
	// put this in database

	return
}

// GetRegistrationStringArgs holds the args for register
type GetRegistrationStringArgs struct {
	// empty
}

// GetRegistrationStringReply holds the data for the register reply
type GetRegistrationStringReply struct {
	RegistrationString string
}

// GetRegistrationString returns a string to the client which is a valid string to sign to indicate they want their pubkey to be registered. This is like kinda weird but whatever
func (cl *OpencxRPC) GetRegistrationString(args GetRegistrationStringArgs, reply *GetRegistrationStringReply) (err error) {
	reply.RegistrationString = cl.Server.GetRegistrationString()
	return
}
