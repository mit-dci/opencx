package cxrpc

import (
	"fmt"
)

// RegisterArgs holds the args for register
type RegisterArgs struct {
	Username string
	Password string
}

// RegisterReply holds the data for the reply
type RegisterReply struct {
	Token []byte
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

	fmt.Println("recieved register")
	// put this in database
	reply.Token = []byte("sampleToken")

	return nil
}
