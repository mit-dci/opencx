package cxdbmemory

import (
	"fmt"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// PlaceAuctionPuzzle puts a puzzle and ciphertext in the datastore.
func (db *CXDBMemory) PlaceAuctionPuzzle(encryptedOrder *match.EncryptedAuctionOrder) (err error) {

	db.puzzleMtx.Lock()
	db.puzzles[encryptedOrder.IntendedAuction] = append(db.puzzles[encryptedOrder.IntendedAuction], encryptedOrder)
	db.puzzleMtx.Unlock()
	return
}

// PlaceAuctionOrder places an order in the unencrypted datastore.
func (db *CXDBMemory) PlaceAuctionOrder(order *match.AuctionOrder) (err error) {

	// TODO: Determine where order validation should go if not here
	// What determines a valid order should be in one place
	if _, err = order.Price(); err != nil {
		err = fmt.Errorf("No price can be determined, so invalid order")
		return
	}

	db.ordersMtx.Lock()
	db.orders[order.AuctionID] = append(db.orders[order.AuctionID], order)
	db.ordersMtx.Unlock()
	return
}

// ViewAuctionOrderBook takes in a trading pair and auction ID, and returns auction orders.
func (db *CXDBMemory) ViewAuctionOrderBook(tradingPair *match.Pair, auctionID [32]byte) (book map[float64][]*match.AuctionOrderIDPair, err error) {

	db.ordersMtx.Lock()
	var allOrders []*match.AuctionOrder
	var ok bool
	if allOrders, ok = db.orders[auctionID]; !ok {
		db.ordersMtx.Unlock()
		err = fmt.Errorf("Could not find auctionID in the auction orderbook")
		return
	}
	var orderPrice float64
	var thisOrderPair *match.AuctionOrderIDPair
	for _, order := range allOrders {
		if order.TradingPair == *tradingPair {
			if orderPrice, err = order.Price(); err != nil {
				db.ordersMtx.Unlock()
				err = fmt.Errorf("Error getting price for order while viewing auction orderbook: %s", err)
				return
			}

			// get hash of order lol
			hasher := sha3.New256()
			hasher.Write(order.SerializeSignable())
			thisOrderPair = new(match.AuctionOrderIDPair)
			copy(thisOrderPair.OrderID[:], hasher.Sum(nil))
			book[orderPrice] = append(book[orderPrice], thisOrderPair)
		}
	}

	db.ordersMtx.Unlock()
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted. This also doesn't error out because if there are no orders with the auctionID
// then it doesn't really matter, we just return an empty list? Maybe we should have an API method
// to return auctionIDs.
func (db *CXDBMemory) ViewAuctionPuzzleBook(auctionID [32]byte) (orders []*match.EncryptedAuctionOrder, err error) {

	db.puzzleMtx.Lock()
	var ok bool
	if orders, ok = db.puzzles[auctionID]; !ok {
		logging.Debugf("Tried to find puzzles matching %x but none were found in DB.", auctionID)
	}
	db.puzzleMtx.Unlock()
	return
}

// NewAuction takes in an auction ID, and creates a new auction, returning the "height"
// of the auction.
func (db *CXDBMemory) NewAuction(auctionID [32]byte) (height uint64, err error) {
	// TODO
	return
}
