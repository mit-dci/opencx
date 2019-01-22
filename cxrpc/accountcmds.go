package cxrpc

import (
	"fmt"
)

// RegisterArgs holds the args for register
type RegisterArgs struct {
	username string
	password string
}

// RegisterReply holds the data for the reply
type RegisterReply struct {
	token []byte
}

// Register registers the user for an account with a username and password
func(cl *OpencxRPC) Register(args RegisterArgs, reply *RegisterReply) error {

	// check for username in database
	// if yes return error

	// TODO: get a database working for user account info and token holding

	// delete this after you get a database working
	inDatabase := false
	if inDatabase {
		return fmt.Errorf("username in database")
	}

	// put this in database
	reply.token = []byte("sampleToken")

	return nil
}
