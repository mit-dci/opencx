package main

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
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

	cl.Printf("Balance for token %s: %d\n", balanceArgs.Asset, balanceReply.Amount)
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

	cl.Printf("DepositAddress for token %s: %s\n", depositArgs.Asset, depositReply.Address)
	return nil
}
