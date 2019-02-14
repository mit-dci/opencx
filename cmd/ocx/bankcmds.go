package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

func (cl *openCxClient) GetBalance(args []string) error {
	balanceArgs := new(cxrpc.GetBalanceArgs)
	balanceReply := new(cxrpc.GetBalanceReply)

	username := args[0]
	asset := args[1]

	balanceArgs.Username = username
	balanceArgs.Asset = asset

	err := cl.Call("OpencxRPC.GetBalance", balanceArgs, balanceReply)
	if err != nil {
		return fmt.Errorf("Error calling 'GetBalance' service method:\n%s", err)
	}

	logging.Infof("Balance for token %s: %f %s\n", balanceArgs.Asset, float64(balanceReply.Amount)/math.Pow10(8), balanceArgs.Asset)
	return nil
}

func (cl *openCxClient) GetDepositAddress(args []string) error {
	depositArgs := new(cxrpc.GetDepositAddressArgs)
	depositReply := new(cxrpc.GetDepositAddressReply)

	username := args[0]
	asset := args[1]

	depositArgs.Username = username
	depositArgs.Asset = asset

	err := cl.Call("OpencxRPC.GetDepositAddress", depositArgs, depositReply)
	if err != nil {
		return fmt.Errorf("Error calling 'GetDepositAddress' service method:\n%s", err)
	}

	logging.Infof("DepositAddress for token %s: %s\n", depositArgs.Asset, depositReply.Address)
	return nil
}

// GetAllBalances get the balance for every token
func (cl *openCxClient) GetAllBalances(args []string) error {
	var err error

	err = cl.GetBalance([]string{args[0], "btc"})
	if err != nil {
		return err
	}

	err = cl.GetBalance([]string{args[0], "vtc"})
	if err != nil {
		return err
	}

	err = cl.GetBalance([]string{args[0], "ltc"})
	if err != nil {
		return err
	}

	return nil
}

func (cl *openCxClient) Withdraw(args []string) error {
	var err error

	withdrawArgs := new(cxrpc.WithdrawArgs)
	withdrawReply := new(cxrpc.WithdrawReply)

	withdrawArgs.Username = args[0]
	withdrawArgs.Amount, err = strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return err
	}
	withdrawArgs.Asset = args[2]
	withdrawArgs.Address = args[3]

	if err := cl.Call("OpencxRPC.Withdraw", withdrawArgs, withdrawReply); err != nil {
		return fmt.Errorf("Error calling 'Withdraw' service method:\n%s", err)
	}

	if withdrawReply.Txid == "" {
		return fmt.Errorf("Error: Unsupported Asset")
	}

	logging.Infof("Withdraw transaction ID: %s\n", withdrawReply.Txid)

	return nil
}
