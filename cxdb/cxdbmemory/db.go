package cxdbmemory

import (
	"sync"

	"github.com/Rjected/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

// CXDBMemory is a super basic data store that just uses golang types for everything
// there's no persistence, it's just used to build out the outer layers of a feature
// before the persistent database details are worked out
type CXDBMemory struct {
	balances    map[*pubkeyCoinPair]uint64
	balancesMtx *sync.Mutex
	puzzles     map[[32]byte][]*match.EncryptedAuctionOrder
	puzzleMtx   *sync.Mutex
	orders      map[[32]byte][]*match.AuctionOrder
	ordersMtx   *sync.Mutex
}

type pubkeyCoinPair struct {
	pubkey [33]byte
	coin   *coinparam.Params
}

// SetupClient makes sure that whatever things need to be done before we use the datastore can be done before we need to use the datastore.
func (db *CXDBMemory) SetupClient(coins []*coinparam.Params) (err error) {

	db.puzzles = make(map[[32]byte][]*match.EncryptedAuctionOrder)
	db.puzzleMtx = new(sync.Mutex)

	db.balances = make(map[*pubkeyCoinPair]uint64)
	db.balancesMtx = new(sync.Mutex)

	db.orders = make(map[[32]byte][]*match.AuctionOrder)
	db.ordersMtx = new(sync.Mutex)

	return
}
