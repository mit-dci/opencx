package benchclient

import (
	"github.com/mit-dci/opencx/cxrpc"
)

// Register registers the user for an account with a username and password
func (cl *BenchClient) Register(username string) (registerReply *cxrpc.RegisterReply, err error) {
	registerReply = new(cxrpc.RegisterReply)
	registerArgs := &cxrpc.RegisterArgs{
		Username: username,
	}

	if err = cl.Call("OpencxRPC.Register", registerArgs, registerReply); err != nil {
		return
	}

	return
}
