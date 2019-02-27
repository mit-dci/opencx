package benchclient

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

// Register registers the user for an account with a username and password
func (cl *BenchClient) Register(username string) (registerReply *cxrpc.RegisterReply, err error) {

	registerArgs := new(cxrpc.RegisterArgs)

	registerArgs.Username = username

	if err = cl.Call("OpencxRPC.Register", registerArgs, registerReply); err != nil {
		err = fmt.Errorf("Error calling 'Register' service method:\n%s", err)
	}

	return
}
