package cxdbmemory

import (
	"fmt"

	"github.com/mit-dci/opencx/match"
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
func (db *CXDBMemory) ViewAuctionOrderBook(tradingPair *match.Pair, auctionID [32]byte) (sellOrderBook []*match.AuctionOrder, buyOrderBook []*match.AuctionOrder, err error) {

	db.ordersMtx.Lock()
	var allOrders []*match.AuctionOrder
	var ok bool
	if allOrders, ok = db.orders[auctionID]; !ok {
		db.ordersMtx.Unlock()
		err = fmt.Errorf("Could not find auctionID in the auction orderbook")
		return
	}
	for _, order := range allOrders {
		if order.TradingPair == *tradingPair {
			if order.IsBuySide() {
				buyOrderBook = append(buyOrderBook, order)
			} else if order.IsSellSide() {
				sellOrderBook = append(sellOrderBook, order)
			}
		}
	}

	db.ordersMtx.Unlock()
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted.
func (db *CXDBMemory) ViewAuctionPuzzleBook(auctionID [32]byte) (orders []*match.EncryptedAuctionOrder, err error) {

	db.puzzleMtx.Lock()
	var ok bool
	if orders, ok = db.puzzles[auctionID]; !ok {
		db.puzzleMtx.Unlock()
		err = fmt.Errorf("Could not find auctionID in the puzzle orderbook")
		return
	}
	db.puzzleMtx.Unlock()
	return
}
