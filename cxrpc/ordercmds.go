package cxrpc

import (
	"github.com/mit-dci/opencx/match"
)

// SubmitOrderArgs holds the args for the submitorder command
type SubmitOrderArgs struct {
	BuyOrder  *match.Order
	SellOrder *match.Order
}

// SubmitOrderReply holds the args for the submitorder command
type SubmitOrderReply struct {
	// TODO empty for now
}

// SubmitOrder submits an order to the order book or throws an error
func(cl *OpencxRPC) SubmitOrder(args SubmitOrderArgs, reply *SubmitOrderReply) error {
	return nil
}
