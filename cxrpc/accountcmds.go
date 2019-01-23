package cxrpc

import (
	"fmt"
)

// RegisterArgs holds the args for register
type RegisterArgs struct {
	Username string
	Password string
}

// RegisterReply holds the data for the register reply
type RegisterReply struct {
	Token []byte
}

// LoginArgs holds the args for login
type LoginArgs struct {
	Username string
	Password string
}

// LoginReply holds the data for the login reply
type LoginReply struct {
	Token []byte
}

// Register registers the user for an account with a username and password
func(cl *OpencxRPC) Register(args RegisterArgs, reply *RegisterReply) error {

	// check for username in database
	// if yes return error

	// TODO: get a database working for user account info and token holding
	// TODO: make a generic CreateAndStoreToken(username) function
	// TODO: make a generic CheckToken(username) function

	// delete this after you get a database working
	inDatabase := false
	if inDatabase {
		return fmt.Errorf("username in database")
	}

	fmt.Printf("Registering user with username %s\n", args.Username)
	// put this in database
	reply.Token = []byte("sampleToken")

	return nil
}

// Login checks the username and password, and sends the user a token
func(cl *OpencxRPC) Login(args LoginArgs, reply *LoginReply) error {
	// TODO
	return nil
}
