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

// HandleForAuction tries to receive all solved auction orders for a specific auction, batches them up, and then places the batch.
// chanSize is the number of orders that should be in this auction.
func (s *OpencxAuctionServer) HandleForAuction(auctionResultChannel chan *match.OrderPuzzleResult, chanSize uint64) {
	// We can reuse these, do not put them in the infinite loop
	var receivedOrder *match.OrderPuzzleResult
	var err error
	var orderBatch []*match.AuctionOrder
	for ; chanSize > 0; chanSize-- {
		receivedOrder = <-auctionResultChannel
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
		orderBatch = append(orderBatch, receivedOrder.Auction)
	}
}
