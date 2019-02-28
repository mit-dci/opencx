package benchclient

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

// GetBalance calls the getbalance rpc command
func (cl *BenchClient) GetBalance(username string, asset string) (getBalanceReply *cxrpc.GetBalanceReply, err error) {
	getBalanceReply = new(cxrpc.GetBalanceReply)
	getBalanceArgs := &cxrpc.GetBalanceArgs{
		Username: username,
		Asset:    asset,
	}

	if err = cl.Call("OpencxRPC.GetBalance", getBalanceArgs, getBalanceReply); err != nil {
		return
	}

	return
}

// GetDepositAddress calls the getdepositaddress rpc command
func (cl *BenchClient) GetDepositAddress(username string, asset string) (getDepositAddressReply *cxrpc.GetDepositAddressReply, err error) {
	getDepositAddressReply = new(cxrpc.GetDepositAddressReply)
	getDepositAddressArgs := &cxrpc.GetDepositAddressArgs{
		Username: username,
		Asset:    asset,
	}

	if err = cl.Call("OpencxRPC.GetDepositAddress", getDepositAddressArgs, getDepositAddressReply); err != nil {
		return
	}

	return
}

// GetAllBalances get the balance for every token
func (cl *BenchClient) GetAllBalances(username string) (balances map[string]uint64, err error) {
	var reply *cxrpc.GetBalanceReply
	if reply, err = cl.GetBalance(username, "btc"); err != nil {
		return
	}

	balances["btc"] = reply.Amount

	if reply, err = cl.GetBalance(username, "vtc"); err != nil {
		return
	}
	balances["vtc"] = reply.Amount

	if reply, err = cl.GetBalance(username, "ltc"); err != nil {
		return
	}
	balances["ltc"] = reply.Amount

	return
}

// Withdraw calls the withdraw rpc command
func (cl *BenchClient) Withdraw(username string, amount uint64, asset string, address string) (withdrawReply *cxrpc.WithdrawReply, err error) {
	withdrawReply = new(cxrpc.WithdrawReply)
	withdrawArgs := &cxrpc.WithdrawArgs{
		Username: username,
		Amount:   amount,
		Asset:    asset,
		Address:  address,
	}

	if err = cl.Call("OpencxRPC.Withdraw", withdrawArgs, withdrawReply); err != nil {
		return
	}

	if withdrawReply.Txid == "" {
		err = fmt.Errorf("Error: Unsupported Asset")
		return
	}

	return
}
