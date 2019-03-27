package cxrpc

import (
	"fmt"

	"github.com/mit-dci/opencx/util"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// GetBalanceArgs hold the arguments for GetBalance
type GetBalanceArgs struct {
	Asset     string
	Signature []byte
}

// GetBalanceReply holds the reply for GetBalance
type GetBalanceReply struct {
	Amount uint64
}

// GetBalance is the RPC Interface for GetBalance
func (cl *OpencxRPC) GetBalance(args GetBalanceArgs, reply *GetBalanceReply) (err error) {

	// e = h(asset)
	sha3 := sha3.New256()
	sha3.Write([]byte(args.Asset))
	e := sha3.Sum(nil)

	pubkey, _, err := koblitz.RecoverCompact(koblitz.S256(), args.Signature, e)
	if err != nil {
		err = fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
		return
	}

	cl.Server.LockIngests()
	if reply.Amount, err = cl.Server.OpencxDB.GetBalance(pubkey, args.Asset); err != nil {
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error with getbalance command: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()

	return
}

// GetDepositAddressArgs hold the arguments for GetDepositAddress
type GetDepositAddressArgs struct {
	Asset     string
	Signature []byte
}

// GetDepositAddressReply holds the reply for GetDepositAddress
type GetDepositAddressReply struct {
	Address string
}

// GetDepositAddress is the RPC Interface for GetDepositAddress
func (cl *OpencxRPC) GetDepositAddress(args GetDepositAddressArgs, reply *GetDepositAddressReply) (err error) {

	// e = h(asset)
	sha3 := sha3.New256()
	sha3.Write([]byte(args.Asset))
	e := sha3.Sum(nil)

	pubkey, _, err := koblitz.RecoverCompact(koblitz.S256(), args.Signature, e)
	if err != nil {
		err = fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
		return
	}

	cl.Server.LockIngests()
	if reply.Address, err = cl.Server.OpencxDB.GetDepositAddress(pubkey, args.Asset); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error with getdepositaddress command: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()

	return
}

// WithdrawArgs holds the args for Withdraw
type WithdrawArgs struct {
	Withdrawal *match.Withdrawal
	Signature  []byte
}

// TODO: figure out a good way to do this serialize and signature stuff!!

// WithdrawReply holds the reply for Withdraw
type WithdrawReply struct {
	Txid string
}

// Withdraw is the RPC Interface for Withdraw
func (cl *OpencxRPC) Withdraw(args WithdrawArgs, reply *WithdrawReply) (err error) {

	// e = h(asset)
	sha3 := sha3.New256()
	sha3.Write(args.Withdrawal.Serialize())
	e := sha3.Sum(nil)

	pubkey, _, err := koblitz.RecoverCompact(koblitz.S256(), args.Signature, e)
	if err != nil {
		err = fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
		return
	}

	var coinType *coinparam.Params
	if coinType, err = util.GetParamFromName(args.Withdrawal.Asset.String()); err != nil {
		return
	}

	cl.Server.LockIngests()
	if reply.Txid, err = cl.Server.WithdrawCoins(args.Withdrawal.Address, pubkey, args.Withdrawal.Amount, coinType); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error with withdraw command: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()

	return
}
