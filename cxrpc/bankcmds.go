package cxrpc

import (
	"fmt"
)

// GetBalanceArgs hold the arguments for GetBalance
type GetBalanceArgs struct {
	Username string
	Asset    string
}

// GetBalanceReply holds the reply for GetBalance
type GetBalanceReply struct {
	Amount uint64
}

// GetBalance is the RPC Interface for GetBalance
func(cl *OpencxRPC) GetBalance(args GetBalanceArgs, reply *GetBalanceReply) error {
	amount, err := cl.Server.OpencxDB.GetBalance(args.Username, args.Asset)
	if err != nil {
		return fmt.Errorf("Error with getbalance command: \n%s", err)
	}

	reply.Amount = amount
	return nil
}
