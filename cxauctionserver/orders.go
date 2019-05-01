package cxauctionserver

import (
	"fmt"

	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()
	if err = s.OpencxDB.PlaceAuctionPuzzle(order); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	// send order solving to channel
	go order.SolveRC5AuctionOrderAsync(s.orderChannel)

	return
}

// decryptPlaceOrder is what we call after committing to an order
func (s *OpencxAuctionServer) decryptPlaceOrder(order *match.EncryptedAuctionOrder) (err error) {

	return
}

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateOrder(decryptedOrder *match.AuctionOrder, encryptedOrder *match.EncryptedAuctionOrder) (valid bool, err error) {

	if _, err = decryptedOrder.Price(); err != nil {
		err = fmt.Errorf("Orders with an indeterminable price are invalid: %s", err)
		return
	}

	if !decryptedOrder.IsBuySide() && !decryptedOrder.IsSellSide() {
		err = fmt.Errorf("Orders that aren't buy or sell side are invalid: %s", err)
		return
	}

	return
}
