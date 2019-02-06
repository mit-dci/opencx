package cxrpc

import (
	"github.com/mit-dci/opencx/match"
)

// SubmitOrderArgs holds the args for the submitorder command
type SubmitOrderArgs struct {
	Order *match.LimitOrder
}

// SubmitOrderReply holds the args for the submitorder command
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
