package cxdbmemory

import (
	"fmt"
	"sync"

	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

// MemoryPuzzleStore is a puzzle store representation for an in memory database
type MemoryPuzzleStore struct {
	puzzles   map[match.AuctionID][]*match.EncryptedAuctionOrder
	puzzleMtx *sync.Mutex
	// the pair for this puzzle store
	// this is just for convenience, the protocol still works if you have one massive puzzle store
	// but if you run many markets at once then you may want to invalidate orders that weren't submitted
	// for the pair they said they were
	pair *match.Pair
}

// CreatePuzzleStore creates a puzzle store for a specific coin.
func CreatePuzzleStore(pair *match.Pair) (store cxdb.PuzzleStore, err error) {
	// Set values
	mp := &MemoryPuzzleStore{
		puzzles:   make(map[match.AuctionID][]*match.EncryptedAuctionOrder),
		puzzleMtx: new(sync.Mutex),
		pair:      pair,
	}
	// Now we actually set the engine
	store = mp
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted.
func (mp *MemoryPuzzleStore) ViewAuctionPuzzleBook(auctionID *match.AuctionID) (puzzles []*match.EncryptedAuctionOrder, err error) {
	mp.puzzleMtx.Lock()
	for _, pzList := range mp.puzzles {
		puzzles = append(puzzles, pzList...)
	}
	mp.puzzleMtx.Unlock()
	return
}

// PlaceAuctionPuzzle puts an encrypted auction order in the datastore.
func (mp *MemoryPuzzleStore) PlaceAuctionPuzzle(puzzledOrder *match.EncryptedAuctionOrder) (err error) {
	mp.puzzleMtx.Lock()
	mp.puzzles[puzzledOrder.IntendedAuction] = append(mp.puzzles[puzzledOrder.IntendedAuction], puzzledOrder)
	mp.puzzleMtx.Unlock()
	return
}

// CreatePuzzleStoreMap creates a map of pair to pair list, given a list of pairs.
func CreatePuzzleStoreMap(pairList []*match.Pair) (pzMap map[match.Pair]cxdb.PuzzleStore, err error) {

	pzMap = make(map[match.Pair]cxdb.PuzzleStore)
	var curPzEng cxdb.PuzzleStore
	for _, pair := range pairList {
		if curPzEng, err = CreatePuzzleStore(pair); err != nil {
			err = fmt.Errorf("Error creating single puzzle store while creating puzzle store map: %s", err)
			return
		}
		pzMap[*pair] = curPzEng
	}

	return
}
