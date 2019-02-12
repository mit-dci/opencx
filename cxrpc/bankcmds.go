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
	cl.Server.LockIngests()
	addr, err := cl.Server.OpencxDB.GetDepositAddress(args.Username, args.Asset)
	if err != nil {
		return fmt.Errorf("Error with getdepositaddress command: \n%s", err)
	}
	cl.Server.UnlockIngests()

	reply.Address = addr
	return nil
}

// WithdrawArgs holds the args for Withdraw
type WithdrawArgs struct {
	Username string
	Asset    string
	Amount   uint64
	Address  string
}

// WithdrawReply holds the reply for Withdraw
type WithdrawReply struct {
	Txid string
}

// Withdraw is the RPC Interface for Withdraw
func (cl *OpencxRPC) Withdraw(args WithdrawArgs, reply *WithdrawReply) error {
	if args.Asset == "vtc" {
		cl.Server.LockIngests()
		txid, err := cl.Server.VTCWithdraw(args.Address, args.Username, args.Amount)
		if err != nil {
			return fmt.Errorf("Error with withdraw command: \n%s", err)
		}
		cl.Server.UnlockIngests()

		reply.Txid = txid
		return nil
	}
	if args.Asset == "btc" {
		cl.Server.LockIngests()
		txid, err := cl.Server.BTCWithdraw(args.Address, args.Username, args.Amount)
		if err != nil {
			return fmt.Errorf("Error with withdraw command: \n%s", err)
		}
		cl.Server.UnlockIngests()

		reply.Txid = txid
		return nil
	}
	if args.Asset == "ltc" {
		cl.Server.LockIngests()
		txid, err := cl.Server.LTCWithdraw(args.Address, args.Username, args.Amount)
		if err != nil {
			return fmt.Errorf("Error with withdraw command: \n%s", err)
		}
		cl.Server.UnlockIngests()

		reply.Txid = txid
		return nil
	}

	reply.Txid = ""

	return nil
}
