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
	"golang.org/x/text/number"
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

	// clock off button
	clockOffButton chan bool
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
		clockOffButton:    make(chan bool, 1),
	}

	return
}

// StartAuctionWithID starts an auction for a certain pair with a specific ID
func (s *OpencxAuctionServer) StartAuctionWithID(pair *match.Pair, auctionID [32]byte) (err error) {
	logging.Infof("Starting an auction with auction time %d", s.t)
	// get the batcher
	s.dbLock.Lock()
	var currBatcher match.AuctionBatcher
	var ok bool
	if currBatcher, ok = s.OrderBatchers[*pair]; !ok {
		err = fmt.Errorf("Error getting correct batcher for pair %s", pair.String())
		s.dbLock.Unlock()
		return
	}
	s.dbLock.Unlock()

	if err = currBatcher.RegisterAuction(auctionID); err != nil {
		err = fmt.Errorf("Error registering auction with id for StartAuctionWithID: %s", err)
		return
	}

	return
}

// EndAuctionWithID ends an auction for a certain pair with a specific ID, waiting for the auction to end.
// This is blocking, returning the result of the auction.
func (s *OpencxAuctionServer) EndAuctionWithID(pair *match.Pair, auctionID [32]byte) (result *match.AuctionBatch, err error) {
	logging.Infof("Ending auction %x", auctionID)
	// get the batcher
	s.dbLock.Lock()
	var currBatcher match.AuctionBatcher
	var ok bool
	if currBatcher, ok = s.OrderBatchers[*pair]; !ok {
		err = fmt.Errorf("Error getting correct batcher for pair %s", pair.String())
		s.dbLock.Unlock()
		return
	}
	s.dbLock.Unlock()

	var batchResultChan chan *match.AuctionBatch
	if batchResultChan, err = currBatcher.EndAuction(auctionID); err != nil {
		err = fmt.Errorf("Error ending auction with id for EndAuctionWithID: %s", err)
		return
	}

	result = <-batchResultChan
	logging.Infof("Results for auction %x retrieved", auctionID)
	// get the batcher

	return
}

func (s *OpencxAuctionServer) StartClockRandomAuction() (err error) {

	s.dbLock.Lock()
	var randID [32]byte
	var bytesRead int
	for pair, batcher := range s.OrderBatchers {

		// Set auctionID to something random for each batcher
		if bytesRead, err = rand.Read(randID[:]); err != nil {
			err = fmt.Errorf("Error getting random auction ID for initializing server: %s", err)
			s.dbLock.Unlock()
			return
		}

		logging.Infof("Read %d bytes for auctionID! Starting first auction.", bytesRead)

		// // Start the solved order handler (TODO: is this the right place to put this?)
		batcher.RegisterAuction(randID)

		// Start the auction clock (also TODO: is this the right place to put this?)
		go s.AuctionClock(pair, randID)
	}
	s.dbLock.Unlock()

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
			id = currID
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

// StopClock stops the server clock
func (s *OpencxAuctionServer) StopClock() (err error) {
	s.clockOffButton <- true
	return
}

// StopClockAndWait stops the clock and ends, waits for all active auctions to finish
func (s *OpencxAuctionServer) StopClockAndWait() (err error) {
	// Since we're acquiring this lock, we know other things won't end the auctions
	// in the time that we're trying to

	logging.Infof("Trying to stop clock...")
	if err = s.StopClock(); err != nil {
		err = fmt.Errorf("Error stopping clock for StopClockAndWait: %s", err)
		return
	}

	logging.Infof("waiting for results to roll in...")
	howManyBatchers := float64(len(s.OrderBatchers))
	var activeAuctions map[[32]byte]time.Time
	num := 0
	doneChan := make(chan bool)
	for _, batcher := range s.OrderBatchers {
		activeAuctions = batcher.ActiveAuctions()
		for id, _ := range activeAuctions {
			go func() {
				var batchChan chan *match.AuctionBatch
				if batchChan, err = batcher.EndAuction(id); err != nil {
					err = fmt.Errorf("Error ending auction for StopClockAndWait: %s", err)
					return
				}
				<-batchChan
				close(batchChan)
				doneChan <- true
			}()
			num++
		}
		logging.Infof("Ending auction results progress: %v ", number.Percent(float64(num)/howManyBatchers))
	}

	for i := num; i > 0; i-- {
		<-doneChan
	}

	return
}

// CurrentAuctionTime gets the current auction time
func (s *OpencxAuctionServer) CurrentAuctionTime() (currentAuctionTime uint64, err error) {
	currentAuctionTime = s.t
	return
}
