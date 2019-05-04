package cxauctionrpc

import (
	"fmt"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// SubmitPuzzledOrderArgs holds the args for the submitpuzzledorder command
type SubmitPuzzledOrderArgs struct {
	// Use the serialize method on match.EncryptedAuctionOrder
	EncryptedOrderBytes []byte
}

// SubmitPuzzledOrderReply holds the reply for the submitpuzzledorder command
type SubmitPuzzledOrderReply struct {
	// empty
}

// SubmitPuzzledOrder submits an order to the order book or throws an error
func (cl *OpencxAuctionRPC) SubmitPuzzledOrder(args SubmitPuzzledOrderArgs, reply *SubmitPuzzledOrderReply) (err error) {

	logging.Infof("Received timelocked order!")

	order := new(match.EncryptedAuctionOrder)
	if err = order.Deserialize(args.EncryptedOrderBytes); err != nil {
		err = fmt.Errorf("Error deserializing puzzled order: %s", err)
		return
	}

	if err = cl.Server.PlacePuzzledOrder(order); err != nil {
		err = fmt.Errorf("Error placing order while submitting order: \n%s", err)
		return
	}

	return
}
