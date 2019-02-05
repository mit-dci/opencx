package main

import (
	"fmt"
	"strconv"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/match"
)

// OrderCommand submits an order (for now) TODO
func (cl *openCxClient) OrderCommand(args []string) error {
	orderArgs := new(cxrpc.SubmitOrderArgs)
	orderReply := new(cxrpc.SubmitOrderReply)

	var newOrder match.LimitOrder
	newOrder.Client = args[0]
	newOrder.Side = args[1]
	if newOrder.Side != "buy" && newOrder.Side != "sell" {
		return fmt.Errorf("Order's side isn't buy or sell, try again")
	}
	err := newOrder.TradingPair.FromString(args[2])
	if err != nil {
		return fmt.Errorf("Error getting asset pair from string: \n%s", err)
	}
	newOrder.AmountHave, err = strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing amountHave, please enter something valid:\n%s", err)
	}
	price, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return fmt.Errorf("Error parsing price: \n%s", err)
	}
	newOrder.SetAmountWant(price)

	orderArgs.Order = &newOrder
	err = cl.Call("OpencxRPC.SubmitOrder", orderArgs, orderReply)
	if err != nil {
		return fmt.Errorf("Error calling 'SubmitOrder' service method:\n%s", err)
	}

	return nil
}

// ViewOrderbook return s the orderbook TODO
func (cl *openCxClient) ViewOrderbook(args []string) error {
	return nil
}
