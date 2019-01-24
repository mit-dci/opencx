package cxrpc

import (
	"fmt"
)

// AuthArgs holds a token and whatever other args we care about
type AuthArgs struct {
	Username string
	Token []byte
}

// AuthReply holds a success string to show the user on success
type AuthReply struct {
	Success bool
}

// DoAuthenticatedThing does an important, authenticated thing
func (cl *OpencxRPC) DoAuthenticatedThing(args AuthArgs, reply *AuthReply) error {
	var err error
	reply.Success, err = cl.Server.OpencxDB.CheckToken(args.Username, args.Token)
	if err != nil {
		return fmt.Errorf("Error checking token: \n%s", err)
	}

	return nil
}
