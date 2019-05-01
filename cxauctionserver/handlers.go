package cxauctionserver

import (
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// AuctionOrderHandler tries to receive solved orders and when it does, it validates them and sends them to be processed
func (s *OpencxAuctionServer) AuctionOrderHandler(orderResultChannel chan *match.OrderPuzzleResult) {
	for {
		receivedOrder := <-orderResultChannel
		if receivedOrder.Err != nil {
			logging.Errorf("Error came in with order solving result: %s", receivedOrder.Err)
			// if there was an error, don't process the order
			continue
		}

		var isOrderValid bool
		var err error
		if isOrderValid, err = s.validateOrder(receivedOrder.Auction, receivedOrder.Encrypted); err != nil {
			logging.Errorf("Error validating order: %s", err)
		} else if !isOrderValid {
			logging.Warnf("Invalid order received")
		}

	}
}
