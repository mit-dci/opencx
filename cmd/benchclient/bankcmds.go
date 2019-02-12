package benchclient

import (
	"fmt"
	"math"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// GetBalance calls the getbalance rpc command
func (cl *BenchClient) GetBalance(username string, asset string) error {
	balanceArgs := &cxrpc.GetBalanceArgs{
		Username: username,
		Asset:    asset,
	}
	balanceReply := new(cxrpc.GetBalanceReply)

	err := cl.Call("OpencxRPC.GetBalance", balanceArgs, balanceReply)
	if err != nil {
		return fmt.Errorf("Error calling 'GetBalance' service method:\n%s", err)
	}

	logging.Infof("Balance for token %s: %f %s\n", balanceArgs.Asset, float64(balanceReply.Amount)/math.Pow10(8), balanceArgs.Asset)
	return nil
}

// GetDepositAddress calls the getdepositaddress rpc command
func (cl *BenchClient) GetDepositAddress(username string, asset string) error {
	depositArgs := &cxrpc.GetDepositAddressArgs{
		Username: username,
		Asset:    asset,
	}
	depositReply := new(cxrpc.GetDepositAddressReply)

	err := cl.Call("OpencxRPC.GetDepositAddress", depositArgs, depositReply)
	if err != nil {
		return fmt.Errorf("Error calling 'GetDepositAddress' service method:\n%s", err)
	}

	logging.Infof("DepositAddress for token %s: %s\n", depositArgs.Asset, depositReply.Address)
	return nil
}

// GetAllBalances get the balance for every token
func (cl *BenchClient) GetAllBalances(username string) (err error) {
	if err = cl.GetBalance(username, "btc"); err != nil {
		return
	}

	if err = cl.GetBalance(username, "vtc"); err != nil {
		return
	}

	if err = cl.GetBalance(username, "ltc"); err != nil {
		return
	}

	return
}

// Withdraw calls the withdraw rpc command
func (cl *BenchClient) Withdraw(username string, amount uint64, asset string, address string) (err error) {
	withdrawArgs := &cxrpc.WithdrawArgs{
		Username: username,
		Amount:   amount,
		Asset:    asset,
		Address:  address,
	}
	withdrawReply := new(cxrpc.WithdrawReply)

	if err = cl.Call("OpencxRPC.Withdraw", withdrawArgs, withdrawReply); err != nil {
		return fmt.Errorf("Error calling 'Withdraw' service method:\n%s", err)
	}

	if withdrawReply.Txid == "" {
		return fmt.Errorf("Error: Unsupported Asset")
	}

	logging.Infof("Withdraw transaction ID: %s\n", withdrawReply.Txid)

	return nil
}
