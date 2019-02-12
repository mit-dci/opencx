package benchclient

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// Register registers the user for an account with a username and password
func (cl *BenchClient) Register(username string) error {

	registerArgs := new(cxrpc.RegisterArgs)
	registerReply := new(cxrpc.RegisterReply)

	registerArgs.Username = username

	err := cl.Call("OpencxRPC.Register", registerArgs, registerReply)
	if err != nil {
		return fmt.Errorf("Error calling 'Register' service method:\n%s", err)
	}

	logging.Infof("Successfully registered\n")
	return nil
}
