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
func (cl *OpencxRPC) Register(args RegisterArgs, reply *RegisterReply) error {

	// check for username in database
	// if yes return error

	// TODO: make a generic CreateAndStoreToken(username) function
	// TODO: make a generic CheckToken(username) function

	success, err := cl.Server.OpencxDB.CreateAccount(args.Username, args.Password)
	if err != nil {
		return fmt.Errorf("Error creating account: \n%s", err)
	}

	// delete this after you get a database working
	if !success {
		reply.Token = nil
		fmt.Printf("Username %s already exists\n", args.Username)
		return nil
	}

	fmt.Printf("Registering user with username %s\n", args.Username)
	// put this in database

	// TODO: once tokens are implemented, remove this
	reply.Token = []byte("sampleToken")

	return nil
}

// Login checks the username and password, and sends the user a token
func (cl *OpencxRPC) Login(args LoginArgs, reply *LoginReply) error {
	success, err := cl.Server.OpencxDB.CheckCredentials(args.Username, args.Password)
	if err != nil {
		return fmt.Errorf("Error logging in: \n%s", err)
	}

	if !success {
		return fmt.Errorf("Credentials incorrect or username doesn't exist")
	}

	reply.Token = []byte("sampleToken")
	return nil
}
