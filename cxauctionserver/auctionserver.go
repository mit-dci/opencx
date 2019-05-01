package cxauctionserver

import (
	"sync"

	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

// OpencxAuctionServer is what will hopefully help handle and manage the auction logic, rpc, and db
type OpencxAuctionServer struct {
	OpencxDB     cxdb.OpencxAuctionStore
	dbLock       *sync.Mutex
	orderChannel chan *match.OrderPuzzleResult
}

// InitServer creates a new server
func InitServer(db cxdb.OpencxAuctionStore, orderChanSize uint64) (server *OpencxAuctionServer) {
	server = &OpencxAuctionServer{
		OpencxDB:     db,
		dbLock:       new(sync.Mutex),
		orderChannel: make(chan *match.OrderPuzzleResult, orderChanSize),
	}
	return
}
