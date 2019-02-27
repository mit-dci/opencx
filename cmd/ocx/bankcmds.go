package main

import (
	"math"
	"strconv"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

func (cl *openCxClient) GetBalance(args []string) (err error) {
	username := args[0]
	asset := args[1]

	var balanceReply *cxrpc.GetBalanceReply
	if balanceReply, err = cl.RPCClient.GetBalance(username, asset); err != nil {
		return
	}

	logging.Infof("Balance for token %s: %f %s\n", asset, float64(balanceReply.Amount)/math.Pow10(8), asset)
	return
}

func (cl *openCxClient) GetDepositAddress(args []string) (err error) {
	username := args[0]
	asset := args[1]

	var getDepositAddressReply *cxrpc.GetDepositAddressReply
	if getDepositAddressReply, err = cl.RPCClient.GetDepositAddress(username, asset); err != nil {
		return
	}

	logging.Infof("DepositAddress for token %s: %s\n", asset, getDepositAddressReply.Address)
	return
}

// GetAllBalances get the balance for every token
func (cl *openCxClient) GetAllBalances(args []string) (err error) {
	username := args[0]

	var getAllBalancesReply map[string]uint64
	if getAllBalancesReply, err = cl.RPCClient.GetAllBalances(username); err != nil {
		return
	}

	for asset, amount := range getAllBalancesReply {
		logging.Infof("Balance for token %s: %f %s\n", asset, float64(amount)/math.Pow10(8), asset)
	}

	return
}

func (cl *openCxClient) Withdraw(args []string) (err error) {
	username := args[0]

	var amount uint64
	if amount, err = strconv.ParseUint(args[1], 10, 64); err != nil {
		return
	}

	asset := args[2]
	address := args[3]

	var withdrawReply *cxrpc.WithdrawReply
	if withdrawReply, err = cl.RPCClient.Withdraw(username, amount, asset, address); err != nil {
		return
	}

	logging.Infof("Withdraw transaction ID: %s\n", withdrawReply.Txid)
	return
}
