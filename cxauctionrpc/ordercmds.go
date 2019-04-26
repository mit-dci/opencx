package cxauctionrpc

import (
	"fmt"

	"github.com/mit-dci/opencx/match"
)

// SubmitPuzzledOrderArgs holds the args for the submitpuzzledorder command
type SubmitPuzzledOrderArgs struct {
	Order *match.EncryptedAuctionOrder
}

// SubmitPuzzledOrderReply holds the reply for the submitpuzzledorder command
type SubmitPuzzledOrderReply struct{}

// SubmitPuzzledOrder submits an order to the order book or throws an error
func (cl *OpencxAuctionRPC) SubmitPuzzledOrder(args SubmitPuzzledOrderArgs, reply *SubmitPuzzledOrderReply) (err error) {

	if err = cl.Server.OpencxDB.PlaceAuctionPuzzle(args.Order.OrderPuzzle, args.Order.OrderCiphertext); err != nil {
		err = fmt.Errorf("Error placing order while submitting order: \n%s", err)
		return
	}

	return
}
