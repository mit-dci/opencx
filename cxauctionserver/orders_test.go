package cxauctionserver

import (
	"testing"
	"time"

	"github.com/mit-dci/opencx/match"
)

// Create constants to be used for tests
var (
	testAuctionOrder = &match.AuctionOrder{
		Nonce:     [2]byte{0xab, 0xcd},
		AuctionID: [32]byte{0xde, 0xad, 0xbe, 0xef},
	}
	testMemoryAmplifier   = uint64(10000)
	testEncryptedOrder, _ = testAuctionOrder.TurnIntoEncryptedOrder(testMemoryAmplifier * testStandardAuctionTime)
	testNumOrders         = 100
	doneChan              = make(chan bool)
)

func TestMemPlacePuzzledOrder(t *testing.T) {
	var err error

	var s *OpencxAuctionServer
	if s, err = initTestServer(); err != nil {
		t.Errorf("Error init test server for TestMemPlacePuzzledOrder: %s", err)
		return
	}

	var orders []*match.EncryptedAuctionOrder
	// Add a bunch of orders to the order list
	for i := 0; i < testNumOrders; i++ {
		orders = append(orders, testEncryptedOrder)
	}

	// Place a bunch of orders
	for _, order := range orders {
		t.Logf("Placing order...\n")
		// We can do this in sequence because it's going to start a goroutine anyways
		s.PlacePuzzledOrder(order)
		// Okay now we'll wait a little bit to make sure we see this memory issue in action
		time.Sleep(time.Duration(testStandardAuctionTime) * 10)
	}

	return
}
