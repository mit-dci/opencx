package cxauctionserver

import (
	"sync"

	"github.com/mit-dci/opencx/cxdb"
)

// OpencxAuctionServer is what will hopefully help handle and manage the auction logic, rpc, and db
type OpencxAuctionServer struct {
	OpencxDB cxdb.OpencxAuctionStore
	dbLock   *sync.Mutex
}
