package cxdbmemory

import (
	"github.com/mit-dci/opencx/crypto"
	"github.com/mit-dci/opencx/match"
)

// PlaceAuctionPuzzle puts a puzzle and ciphertext in the datastore.
func (db *CXDBMemory) PlaceAuctionPuzzle(puzzle crypto.Puzzle, ciphertext []byte) (err error) {

	// TODO
	return
}

// PlaceAuctionOrder places an order in the unencrypted datastore.
func (db *CXDBMemory) PlaceAuctionOrder(*match.AuctionOrder) (err error) {

	// TODO
	return
}

// ViewAuctionOrderBook takes in a trading pair and auction ID, and returns auction orders.
func (db *CXDBMemory) ViewAuctionOrderBook(tradingPair *match.Pair, auctionID [32]byte) (sellOrderBook []*match.AuctionOrder, buyOrderBook []*match.AuctionOrder, err error) {

	// TODO
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted.
func (db *CXDBMemory) ViewAuctionPuzzleBook(auctionID [32]byte) (orders []*match.EncryptedAuctionOrder, err error) {

	// TODO
	return
}
