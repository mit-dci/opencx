package cxauctionserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/cxdb/cxdbmemory"
)

const (
	testOrderChanSize       = 100
	testStandardAuctionTime = 1000000
)

var (
	testCoins = []*coinparam.Params{
		&coinparam.BitcoinParams,
		&coinparam.VertcoinTestNetParams,
	}
)

// initTestServer initializes the server. This is mostly setting up the db
func initTestServer() (s *OpencxAuctionServer, err error) {

	testDB := new(cxdbmemory.CXDBMemory)
	if err = testDB.SetupClient(testCoins); err != nil {
		err = fmt.Errorf("Error setting up db client for tests: %s", err)
		return
	}

	// Initialize the test server
	if s, err = InitServer(testDB, testOrderChanSize, testStandardAuctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for tests: %s", err)
		return
	}

	return
}
