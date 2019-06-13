package cxauctionserver

import (
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// OpencxAuctionServer is what will hopefully help handle and manage the auction logic, rpc, and db
type OpencxAuctionServer struct {
	OpencxDB     cxdb.OpencxAuctionStore
	OrderBatcher match.AuctionBatcher
	dbLock       *sync.Mutex
	orderChannel chan *match.OrderPuzzleResult

	// auction params -- we'll store them in here for now
	auctionID [32]byte
	t         uint64
}

// InitServer creates a new server, with the default batcher ABatcher
func InitServer(db cxdb.OpencxAuctionStore, orderChanSize uint64, standardAuctionTime uint64, maxBatchSize uint64) (server *OpencxAuctionServer, err error) {
	logging.Infof("Starting ABatcher!")
	var batcher *ABatcher
	if batcher, err = NewABatcher(maxBatchSize); err != nil {
		err = fmt.Errorf("Error starting ABatcher while init server: %s", err)
		return
	}

	logging.Infof("Starting an auction with auction time %d", standardAuctionTime)

	server, err = InitServerWithBatcher(db, batcher, orderChanSize, standardAuctionTime)
	return
}

// InitServerWithBatcher creates a new server with a specified batcher
func InitServerWithBatcher(db cxdb.OpencxAuctionStore, batcher match.AuctionBatcher, orderChanSize uint64, standardAuctionTime uint64) (server *OpencxAuctionServer, err error) {
	logging.Infof("Starting an auction with auction time %d", standardAuctionTime)
	server = &OpencxAuctionServer{
		OpencxDB:     db,
		OrderBatcher: batcher,
		dbLock:       new(sync.Mutex),
		orderChannel: make(chan *match.OrderPuzzleResult, orderChanSize),
		t:            standardAuctionTime,
	}

	// Set auctionID to something random
	var bytesRead int
	if bytesRead, err = rand.Read(server.auctionID[:]); err != nil {
		err = fmt.Errorf("Error getting random auction ID for initializing server: %s", err)
		return
	}

	logging.Infof("Read %d bytes for auctionID! Starting first auction.", bytesRead)

	// // Start the solved order handler (TODO: is this the right place to put this?)
	// go server.AuctionOrderHandler(server.orderChannel)
	batcher.RegisterAuction(server.auctionID)

	// Start the auction clock (also TODO: is this the right place to put this?)
	go server.AuctionClock()

	return
}

// CurrentAuctionID gets the current auction ID
func (s *OpencxAuctionServer) CurrentAuctionID() (currentAuctionID [32]byte, err error) {
	currentAuctionID = s.auctionID
	return
}

// CurrentAuctionTime gets the current auction time
func (s *OpencxAuctionServer) CurrentAuctionTime() (currentAuctionTime uint64, err error) {
	currentAuctionTime = s.t
	return
}
