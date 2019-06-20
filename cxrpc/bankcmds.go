package cxrpc

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	util "github.com/mit-dci/opencx/chainutils"
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

	var pubkey *koblitz.PublicKey
	if pubkey, _, err = koblitz.RecoverCompact(koblitz.S256(), args.Signature, e); err != nil {
		err = fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
		return
	}

	var param *coinparam.Params
	if param, err = util.GetParamFromName(args.Asset); err != nil {
		err = fmt.Errorf("Error getting coin type from name, pass in a different asset")
		return
	}

	if reply.Amount, err = cl.Server.GetBalance(pubkey, param); err != nil {
		err = fmt.Errorf("Error getting balance for pubkey in GetBalance RPC command: %s", err)
		return
	}

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

	var pubkey *koblitz.PublicKey
	if pubkey, _, err = koblitz.RecoverCompact(koblitz.S256(), args.Signature, e); err != nil {
		err = fmt.Errorf("Error, invalid signature with GetDepositAddress RPC command: %s", err)
		return
	}

	var param *coinparam.Params
	if param, err = util.GetParamFromName(args.Asset); err != nil {
		err = fmt.Errorf("Error getting param from name for asset: %s", err)
		return
	}

	if reply.Address, err = cl.Server.GetDepositAddress(pubkey, param); err != nil {
		err = fmt.Errorf("Error getting deposit address from server for GetDepositAddress RPC: %s", err)
		return
	}

	return
}

// WithdrawArgs holds the args for Withdraw
type WithdrawArgs struct {
	Withdrawal *match.Withdrawal
	Signature  []byte
}

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

	// We just ignore the address if they specify lightning.

	if args.Withdrawal.Lightning {

		if reply.Txid, err = cl.Server.WithdrawLightning(pubkey, args.Withdrawal.Amount, coinType); err != nil {
			err = fmt.Errorf("Error with withdraw command (withdraw from lightning): \n%s", err)
			return
		}

	} else {

		if reply.Txid, err = cl.Server.WithdrawCoins(args.Withdrawal.Address, pubkey, args.Withdrawal.Amount, coinType); err != nil {
			err = fmt.Errorf("Error with withdraw command (withdraw from chain): \n%s", err)
			return
		}

	}

	return
}
