package cxauctionserver

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbmemory"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// OpencxAuctionServer is what will hopefully help handle and manage the auction logic, rpc, and db
type OpencxAuctionServer struct {
	SettlementEngines map[*coinparam.Params]match.SettlementEngine
	MatchingEngines   map[match.Pair]match.AuctionEngine
	Orderbooks        map[match.Pair]match.AuctionOrderbook
	PuzzleEngines     map[match.Pair]cxdb.PuzzleStore
	OrderBatchers     map[match.Pair]match.AuctionBatcher
	dbLock            *sync.Mutex
	orderChannel      chan *match.OrderPuzzleResult
	orderChanMap      map[[32]byte]chan *match.OrderPuzzleResult

	// auction params -- we'll store them in here for now
	t uint64
}

// InitServerMemoryDefault initializes an auction server with in memory auction engines, settlement engines,
// and puzzle stores
func InitServerMemoryDefault(coinList []*coinparam.Params, orderChanSize uint64, standardAuctionTime uint64, maxBatchSize uint64) (server *OpencxAuctionServer, err error) {

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list for InitServerMemoryDefault: %s", err)
		return
	}

	var mengines map[match.Pair]match.AuctionEngine
	if mengines, err = cxdbmemory.CreateAuctionEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction engine map with coinlist for InitServerMemoryDefault: %s", err)
		return
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbmemory.CreateSettlementEngineMap(coinList); err != nil {
		err = fmt.Errorf("Error creating settlement engine map for InitServerMemoryDefault: %s", err)
		return
	}

	var aucBooks map[match.Pair]match.AuctionOrderbook
	if aucBooks, err = cxdbmemory.CreateAuctionOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction orderbook map for InitServerMemoryDefault: %s", err)
		return
	}

	var pzEngines map[match.Pair]cxdb.PuzzleStore
	if pzEngines, err = cxdbmemory.CreatePuzzleStoreMap(pairList); err != nil {
		err = fmt.Errorf("Error creating puzzle store map for InitServerMemoryDefault: %s", err)
		return
	}

	var batchers map[match.Pair]match.AuctionBatcher
	if batchers, err = CreateAuctionBatcherMap(pairList, maxBatchSize); err != nil {
		err = fmt.Errorf("Error creating batcher map for InitServerSQLDefault: %s", err)
		return
	}

	if server, err = InitServer(setEngines, mengines, aucBooks, pzEngines, batchers, orderChanSize, standardAuctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for InitServerMemoryDefault: %s", err)
		return
	}
	return
}

// InitServerSQLDefault initializes an auction server with SQL engines, orderbooks, and puzzle stores.
// This generates everything using built in methods
func InitServerSQLDefault(coinList []*coinparam.Params, orderChanSize uint64, standardAuctionTime uint64, maxBatchSize uint64) (server *OpencxAuctionServer, err error) {

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

	var mengines map[match.Pair]match.AuctionEngine
	if mengines, err = cxdbsql.CreateAuctionEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction engine map with coinlist for InitServerSQLDefault: %s", err)
		return
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbsql.CreateSettlementEngineMap(coinList); err != nil {
		err = fmt.Errorf("Error creating settlement engine map for InitServerSQLDefault: %s", err)
		return
	}

	var aucBooks map[match.Pair]match.AuctionOrderbook
	if aucBooks, err = cxdbsql.CreateAuctionOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction orderbook map for InitServerSQLDefault: %s", err)
		return
	}

	var pzEngines map[match.Pair]cxdb.PuzzleStore
	if pzEngines, err = cxdbsql.CreatePuzzleStoreMap(pairList); err != nil {
		err = fmt.Errorf("Error creating puzzle store map for InitServerSQLDefault: %s", err)
		return
	}

	var batchers map[match.Pair]match.AuctionBatcher
	if batchers, err = CreateAuctionBatcherMap(pairList, maxBatchSize); err != nil {
		err = fmt.Errorf("Error creating batcher map for InitServerSQLDefault: %s", err)
		return
	}

	if server, err = InitServer(setEngines, mengines, aucBooks, pzEngines, batchers, orderChanSize, standardAuctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}
	return
}

// InitServer creates a new server
func InitServer(setEngines map[*coinparam.Params]match.SettlementEngine, matchEngines map[match.Pair]match.AuctionEngine, books map[match.Pair]match.AuctionOrderbook, pzengines map[match.Pair]cxdb.PuzzleStore, batchers map[match.Pair]match.AuctionBatcher, orderChanSize uint64, standardAuctionTime uint64) (server *OpencxAuctionServer, err error) {
	logging.Infof("Starting an auction with auction time %d", standardAuctionTime)
	server = &OpencxAuctionServer{
		SettlementEngines: setEngines,
		MatchingEngines:   matchEngines,
		Orderbooks:        books,
		PuzzleEngines:     pzengines,
		OrderBatchers:     batchers,
		dbLock:            new(sync.Mutex),
		orderChannel:      make(chan *match.OrderPuzzleResult, orderChanSize),
		orderChanMap:      make(map[[32]byte]chan *match.OrderPuzzleResult),
		t:                 standardAuctionTime,
	}

	server.dbLock.Lock()

	var randID [32]byte
	var bytesRead int
	for pair, batcher := range server.OrderBatchers {

		// Set auctionID to something random for each batcher
		if bytesRead, err = rand.Read(randID[:]); err != nil {
			err = fmt.Errorf("Error getting random auction ID for initializing server: %s", err)
			server.dbLock.Unlock()
			return
		}

		logging.Infof("Read %d bytes for auctionID! Starting first auction.", bytesRead)

		// // Start the solved order handler (TODO: is this the right place to put this?)
		batcher.RegisterAuction(randID)

		// Start the auction clock (also TODO: is this the right place to put this?)
		go server.AuctionClock(pair, randID)
	}
	server.dbLock.Unlock()

	return
}

// GetIDTimeFromPair gets the most recent auction ID and its start time
func (s *OpencxAuctionServer) GetIDTimeFromPair(pair *match.Pair) (id [32]byte, recent time.Time, err error) {

	s.dbLock.Lock()
	var currBatcher match.AuctionBatcher
	var ok bool
	if currBatcher, ok = s.OrderBatchers[*pair]; !ok {
		err = fmt.Errorf("Error getting correct batcher for pair %s", pair.String())
		s.dbLock.Unlock()
		return
	}
	s.dbLock.Unlock()

	// activeauctions has no err because it doesn't really need it
	var activeBatcher map[[32]byte]time.Time = currBatcher.ActiveAuctions()
	// but we make one here if there are no active auctions
	if len(activeBatcher) == 0 {
		err = fmt.Errorf("No active auctions returned, either there actually are no active auctions or it's a bug")
		return
	}

	var start bool = true
	for currID, startTime := range activeBatcher {
		if start {
			// the first iteration, just set the highest to the first time
			// that we see
			recent = startTime
			start = false
			continue
		}

		if recent.Before(startTime) {
			recent = startTime
			id = currID
		}
	}

	// We should have actually set the correct id and time parameters within the loop

	return
}

// CurrentAuctionTime gets the current auction time
func (s *OpencxAuctionServer) CurrentAuctionTime() (currentAuctionTime uint64, err error) {
	currentAuctionTime = s.t
	return
}
