package cxauctionserver

import (
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// AuctionOrderHandler tries to receive solved orders and when it does, it validates them and sends them to be processed
func (s *OpencxAuctionServer) AuctionOrderHandler(orderResultChannel chan *match.OrderPuzzleResult) {
	// We can reuse these, do not put them in the infinite loop
	var receivedOrder *match.OrderPuzzleResult
	var err error
	for {
		receivedOrder = <-orderResultChannel
		if receivedOrder.Err != nil {
			logging.Errorf("Error came in with order solving result: %s", receivedOrder.Err)
			// if there was an error, don't process the order
			continue
		}

		if err = s.validateOrder(receivedOrder.Auction, receivedOrder.Encrypted); err != nil {
			logging.Errorf("Error validating order: %s", err)
			continue
		}

		logging.Infof("Order valid! Order placed by %x", receivedOrder.Auction.Pubkey)

	}
}
