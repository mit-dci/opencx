package main

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

func (cl *openCxClient) GetBalance(args []string) error {
	balanceArgs := new(cxrpc.GetBalanceArgs)
	balanceReply := new (cxrpc.GetBalanceReply)

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
