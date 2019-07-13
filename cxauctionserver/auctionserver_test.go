package cxauctionserver

import (
	"fmt"
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbmemory"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/match"
)

const (
	testOrderChanSize       = uint64(100)
	testStandardAuctionTime = uint64(100000)
	testMaxBatchSize        = uint64(10)
)

var (
	testCoins = []*coinparam.Params{
		&coinparam.BitcoinParams,
		&coinparam.VertcoinTestNetParams,
		&coinparam.RegressionNetParams,
		&coinparam.LiteRegNetParams,
	}
)

// initTestServer initializes the server. This is mostly setting up the db
func initTestServer() (s *OpencxAuctionServer, err error) {

	// Initialize the test server
	if s, err = createUltraLightAuctionServer(testCoins, testOrderChanSize, testStandardAuctionTime, testMaxBatchSize); err != nil {
		err = fmt.Errorf("Error initializing server for tests: %s", err)
		return
	}

	return
}

// createUltraLightAuctionServer creates a server with "pinky swear settlement" after starting the database with a bunch of parameters for everything else
// This one literally has no settlement
func createUltraLightAuctionServer(coinList []*coinparam.Params, orderChanSize uint64, auctionTime uint64, maxBatchSize uint64) (server *OpencxAuctionServer, err error) {

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

	var mengines map[match.Pair]match.AuctionEngine
	if mengines, err = cxdbsql.CreateAuctionEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction engine map with coinlist for createUltraLightAuctionServer: %s", err)
		return
	}

	// These lines are the only difference between the LightAuctionServer and the FullAuctionServer
	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(make(map[*coinparam.Params][][33]byte), true); err != nil {
		err = fmt.Errorf("Error creating pinky swear settlement engine map for createUltraLightAuctionServer: %s", err)
		return
	}

	var aucBooks map[match.Pair]match.AuctionOrderbook
	if aucBooks, err = cxdbsql.CreateAuctionOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction orderbook map for createUltraLightAuctionServer: %s", err)
		return
	}

	var pzEngines map[match.Pair]cxdb.PuzzleStore
	if pzEngines, err = cxdbsql.CreatePuzzleStoreMap(pairList); err != nil {
		err = fmt.Errorf("Error creating puzzle store map for createUltraLightAuctionServer: %s", err)
		return
	}

	var batchers map[match.Pair]match.AuctionBatcher
	if batchers, err = CreateAuctionBatcherMap(pairList, maxBatchSize); err != nil {
		err = fmt.Errorf("Error creating batcher map for createUltraLightAuctionServer: %s", err)
		return
	}

	// orderChanSize = 100 because uh why not?
	if server, err = InitServer(setEngines, mengines, aucBooks, pzEngines, batchers, orderChanSize, auctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for createUltraLightAuctionServer: %s", err)
		return
	}

	return
}

func TestCreateServer(t *testing.T) {
	var err error

	if _, err = initTestServer(); err != nil {
		t.Errorf("Error initializing test server for TestCreateServer: %s", err)
	}

	return
}
