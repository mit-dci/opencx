package cxauctionserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
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
	}
)

// initTestServer initializes the server. This is mostly setting up the db
func initTestServer() (s *OpencxAuctionServer, err error) {

	// Initialize the test server
	if s, err = InitServerMemoryDefault(testCoins, testOrderChanSize, testStandardAuctionTime, testMaxBatchSize); err != nil {
		err = fmt.Errorf("Error initializing server for tests: %s", err)
		return
	}

	return
}
