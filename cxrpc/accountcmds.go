package cxrpc

import (
	"fmt"

	"github.com/mit-dci/opencx/logging"
)

// RegisterArgs holds the args for register
type RegisterArgs struct {
	Username string
}

// RegisterReply holds the data for the register reply
type RegisterReply struct {
	// empty
}

// Register registers the user for an account with a username and password
func (cl *OpencxRPC) Register(args RegisterArgs, reply *RegisterReply) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error registering user: \n%s", err)
		}
	}()

	// Create addresses based on username and put them into maps
	addrMap := make(map[string]string)
	if addrMap["btc"], err = cl.Server.NewAddressBTC(args.Username); err != nil {
		return
	}

	if addrMap["ltc"], err = cl.Server.NewAddressLTC(args.Username); err != nil {
		return
	}

	if addrMap["vtc"], err = cl.Server.NewAddressVTC(args.Username); err != nil {
		return
	}

	// Do all this locking just cause
	cl.Server.LockIngests()
	// Insert them into the DB
	if err = cl.Server.OpencxDB.InsertDepositAddresses(args.Username, addrMap); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		cl.Server.UnlockIngests()
		return
	}
	cl.Server.UnlockIngests()

	cl.Server.LockIngests()
	if err = cl.Server.OpencxDB.InitializeAccountBalances(args.Username); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		cl.Server.UnlockIngests()
		return
	}
	cl.Server.UnlockIngests()

	logging.Infof("Registering user with username %s\n", args.Username)
	// put this in database

	return
}
