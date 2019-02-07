package cxrpc

import (
	"github.com/mit-dci/opencx/match"
)

// SubmitOrderArgs holds the args for the submitorder command
type SubmitOrderArgs struct {
	Order *match.LimitOrder
}

// SubmitOrderReply holds the reply for the submitorder command
type SubmitOrderReply struct {
	// TODO empty for now
}

// SubmitOrder submits an order to the order book or throws an error
func (cl *OpencxRPC) SubmitOrder(args SubmitOrderArgs, reply *SubmitOrderReply) error {

	if err := cl.Server.OpencxDB.PlaceOrder(args.Order); err != nil {
		return err
	}

	if err := cl.Server.OpencxDB.RunMatching(args.Order.TradingPair); err != nil {
		return err
	}

	return nil
}

// ViewOrderBookArgs holds the args for the vieworderbook command
type ViewOrderBookArgs struct {
	TradingPair *match.Pair
}

// ViewOrderBookReply holds the reply for the vieworderbook command
type ViewOrderBookReply struct {
	SellOrderBook []*match.LimitOrder
	BuyOrderBook  []*match.LimitOrder
}

// ViewOrderBook handles the vieworderbook command
func (cl *OpencxRPC) ViewOrderBook(args ViewOrderBookArgs, reply *ViewOrderBookReply) error {
	var err error
	if reply.SellOrderBook, reply.BuyOrderBook, err = cl.Server.OpencxDB.ViewOrderBook(args.TradingPair); err != nil {
		return err
	}

	return nil
}
