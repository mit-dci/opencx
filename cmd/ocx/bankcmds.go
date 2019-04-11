package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mit-dci/lit/lnutil"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

var getBalanceCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.Red("getbalance"), lnutil.ReqColor("asset")),
	Description: fmt.Sprintf("%s\n",
		"Get your balance of asset. You must be registered.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get your balance of asset. You must be registered to run this command."),
}

func (cl *ocxClient) GetBalance(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	asset := args[0]

	var balanceReply *cxrpc.GetBalanceReply
	if balanceReply, err = cl.RPCClient.GetBalance(asset); err != nil {
		return
	}

	logging.Infof("Balance for token %s: %f %s\n", asset, float64(balanceReply.Amount)/math.Pow10(8), asset)
	return
}

var getDepositAddressCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.Red("getdepositaddress"), lnutil.ReqColor("asset")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Get the deposit address for the given asset.",
		"Once you send to this, you will have to wait a certain number of confirmations and then you will be able to trade your coins.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get the deposit address for the given asset."),
}

func (cl *ocxClient) GetDepositAddress(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	asset := args[0]

	var getDepositAddressReply *cxrpc.GetDepositAddressReply
	if getDepositAddressReply, err = cl.RPCClient.GetDepositAddress(asset); err != nil {
		return
	}

	logging.Infof("DepositAddress for token %s: %s\n", asset, getDepositAddressReply.Address)
	return
}

var getAllBalancesCommand = &Command{
	Format: fmt.Sprintf("%s\n", lnutil.Red("getallbalances")),
	Description: fmt.Sprintf("%s\n",
		"Get balances for all tokens supported on the exchange.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get balances for all tokens supported on the exchange."),
}

// GetAllBalances get the balance for every token
func (cl *ocxClient) GetAllBalances(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	var getAllBalancesReply map[string]uint64
	if getAllBalancesReply, err = cl.RPCClient.GetAllBalances(); err != nil {
		return
	}

	for asset, amount := range getAllBalancesReply {
		logging.Infof("Balance for token %s: %f %s\n", asset, float64(amount)/math.Pow10(8), asset)
	}

	return
}

var withdrawCommand = &Command{
	Format: fmt.Sprintf("%s%s%s%s\n", lnutil.Red("withdraw"), lnutil.ReqColor("amount"), lnutil.ReqColor("asset"), lnutil.ReqColor("recvaddress")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Withdraw amount of asset into recvaddress.",
		"Make sure you feel your asset has enough confirmations such that it has been confirmed.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Withdraw amount of asset into recvaddress."),
}

func (cl *ocxClient) Withdraw(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	var amount uint64
	if amount, err = strconv.ParseUint(args[0], 10, 64); err != nil {
		return
	}

	var asset match.Asset
	if asset, err = match.AssetFromString(args[1]); err != nil {
		return
	}
	address := args[2]

	var withdrawReply *cxrpc.WithdrawReply
	if withdrawReply, err = cl.RPCClient.Withdraw(amount, asset, address); err != nil {
		return
	}

	logging.Infof("Withdraw transaction ID: %s\n", withdrawReply.Txid)
	return
}

var litWithdrawCommand = &Command{
	Format: fmt.Sprintf("%s%s%s\n", lnutil.Red("litwithdraw"), lnutil.ReqColor("amount"), lnutil.ReqColor("asset")),
	Description: fmt.Sprintf("%s\n%s\n%s\n",
		"This assumes you're using the same keys as your lightning node, and your node should be running.",
		"This creates a channel with you with the amount you specified.",
		"Make sure you feel your asset has enough confirmations such that it has been confirmed.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Withdraw amount of asset for lightning."),
}

func (cl *ocxClient) LitWithdraw(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	var amount uint64
	if amount, err = strconv.ParseUint(args[0], 10, 64); err != nil {
		return
	}

	var asset match.Asset
	if asset, err = match.AssetFromString(args[1]); err != nil {
		return
	}

	var withdrawReply *cxrpc.WithdrawReply
	if withdrawReply, err = cl.RPCClient.WithdrawLightning(amount, asset); err != nil {
		return
	}

	logging.Infof("Withdraw transaction ID: %s\n", withdrawReply.Txid)
	return
}
