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
func (cl *OpencxRPC) GetBalance(args GetBalanceArgs, reply *GetBalanceReply) error {
	amount, err := cl.Server.OpencxDB.GetBalance(args.Username, args.Asset)
	if err != nil {
		return fmt.Errorf("Error with getbalance command: \n%s", err)
	}

	reply.Amount = amount
	return nil
}

// GetDepositAddressArgs hold the arguments for GetDepositAddress
type GetDepositAddressArgs struct {
	Username string
	Asset    string
}

// GetDepositAddressReply holds the reply for GetDepositAddress
type GetDepositAddressReply struct {
	Address string
}

// GetDepositAddress is the RPC Interface for GetDepositAddress
func (cl *OpencxRPC) GetDepositAddress(args GetDepositAddressArgs, reply *GetDepositAddressReply) error {
	addr, err := cl.Server.OpencxDB.GetDepositAddress(args.Username, args.Asset)
	if err != nil {
		return fmt.Errorf("Error with getdepositaddress command: \n%s", err)
	}

	reply.Address = addr
	return nil
}
