package cxauctionserver

import (
	"crypto/rand"
	"fmt"
	"sync"

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

	dbLock       *sync.Mutex
	orderChannel chan *match.OrderPuzzleResult
	orderChanMap map[[32]byte]chan *match.OrderPuzzleResult

	// auction params -- we'll store them in here for now
	auctionID [32]byte
	t         uint64
}

// InitServerMemoryDefault initializes an auction server with in memory auction engines, settlement engines,
// and puzzle stores
func InitServerMemoryDefault(coinList []*coinparam.Params, orderChanSize uint64, standardAuctionTime uint64) (server *OpencxAuctionServer, err error) {

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

	if server, err = InitServer(setEngines, mengines, aucBooks, pzEngines, orderChanSize, standardAuctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for InitServerMemoryDefault: %s", err)
		return
	}
	return
}

// InitServerSQLDefault initializes an auction server with SQL engines, orderbooks, and puzzle stores.
// This generates everything using built in methods
func InitServerSQLDefault(coinList []*coinparam.Params, orderChanSize uint64, standardAuctionTime uint64) (server *OpencxAuctionServer, err error) {

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

	if server, err = InitServer(setEngines, mengines, aucBooks, pzEngines, orderChanSize, standardAuctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}
	return
}

// InitServer creates a new server
func InitServer(setEngines map[*coinparam.Params]match.SettlementEngine, matchEngines map[match.Pair]match.AuctionEngine, books map[match.Pair]match.AuctionOrderbook, pzengines map[match.Pair]cxdb.PuzzleStore, orderChanSize uint64, standardAuctionTime uint64) (server *OpencxAuctionServer, err error) {
	logging.Infof("Starting an auction with auction time %d", standardAuctionTime)
	server = &OpencxAuctionServer{
		SettlementEngines: setEngines,
		MatchingEngines:   matchEngines,
		Orderbooks:        books,
		PuzzleEngines:     pzengines,
		dbLock:            new(sync.Mutex),
		orderChannel:      make(chan *match.OrderPuzzleResult, orderChanSize),
		orderChanMap:      make(map[[32]byte]chan *match.OrderPuzzleResult),
		t:                 standardAuctionTime,
	}

	// Set auctionID to something random
	if _, err = rand.Read(server.auctionID[:]); err != nil {
		err = fmt.Errorf("Error getting random auction ID for initializing server: %s", err)
		return
	}

	// Start the solved order handler (TODO: is this the right place to put this?)
	go server.AuctionOrderHandler(server.orderChannel)

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
