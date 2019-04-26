package cxauctionserver

import (
	"fmt"

	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	s.dbLock.Lock()
	if err = s.OpencxDB.PlaceAuctionPuzzle(order.OrderPuzzle, order.OrderCiphertext); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	return
}
