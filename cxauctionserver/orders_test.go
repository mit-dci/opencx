package cxauctionserver

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/mit-dci/opencx/match"
)

// Create constants to be used for tests
var (
	testAuctionOrder = &match.AuctionOrder{
		Pubkey:     [...]byte{0x02, 0xe7, 0xb7, 0xcf, 0xcf, 0x42, 0x2f, 0xdb, 0x68, 0x2c, 0x85, 0x02, 0xbf, 0x2e, 0xef, 0x9e, 0x2d, 0x87, 0x67, 0xf6, 0x14, 0x67, 0x41, 0x53, 0x4f, 0x37, 0x94, 0xe1, 0x40, 0xcc, 0xf9, 0xde, 0xb3},
		Nonce:      [2]byte{0x00, 0x00},
		AuctionID:  [32]byte{0xde, 0xad, 0xbe, 0xef},
		AmountWant: 100000,
		AmountHave: 10000,
		Side:       "buy",
		TradingPair: match.Pair{
			AssetWant: match.Asset(6),
			AssetHave: match.Asset(8),
		},
		Signature: []byte{0x1b, 0xd6, 0x0f, 0xd3, 0xec, 0x5b, 0x73, 0xad, 0xa9, 0x8a, 0x92, 0x79, 0x82, 0x0f, 0x8e, 0xab, 0xf8, 0x8f, 0x47, 0x6e, 0xc3, 0x15, 0x33, 0x72, 0xd9, 0x90, 0x51, 0x41, 0xfd, 0x0a, 0xa1, 0xa2, 0x4a, 0x73, 0x75, 0x4c, 0xa5, 0x28, 0x4a, 0xc2, 0xed, 0x5a, 0xe9, 0x33, 0x22, 0xf4, 0x41, 0x1f, 0x9d, 0xd1, 0x78, 0xb9, 0x17, 0xd4, 0xe9, 0x72, 0x51, 0x7f, 0x5b, 0xd7, 0xe5, 0x12, 0xe7, 0x69, 0xb0},
	}
	testEncryptedOrder, _ = testAuctionOrder.TurnIntoEncryptedOrder(testStandardAuctionTime)
	testNumOrders         = 8
	doneChan              = make(chan bool)
)

// incrementNonce increments a 2 byte nonce -- keep in mind that if you use this naively
// in the testAuctionOrder and then timelock puzzle it, it makes the order invalid,
// since the signature is dependent on the nonce.
// We do this in the tests because the auction engine should only match and do no
// other validation.
// However, any outside-facing program SHOULD do other validation and detect this.
func incrementNonce(currNonce [2]byte) (nextNonce [2]byte) {

	var inputSlice [8]byte
	copy(inputSlice[:], currNonce[:])
	currInt, _ := binary.Uvarint(inputSlice[:])
	currInt++
	var outBufSlice []byte = make([]byte, 8)
	binary.PutUvarint(outBufSlice, currInt)
	copy(nextNonce[:], outBufSlice[:1])

	return
}

func TestUltraLightPlacePuzzledOrder(t *testing.T) {
	var err error

	t.Logf("%s: Starting Server", time.Now())
	var s *OpencxAuctionServer
	if s, err = initTestServer(); err != nil {
		t.Errorf("Error init test server for TestMemPlacePuzzledOrder: %s", err)
		return
	}
	t.Logf("%s: Started Server", time.Now())

	t.Logf("%s: Starting Auction", time.Now())
	if err = s.StartAuctionWithID(&testEncryptedOrder.IntendedPair, testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error starting auction with id for TestUltraLightPlacePuzzledOrder: %s", err)
		return
	}
	t.Logf("%s: Started Auction", time.Now())

	baseOrder := *testAuctionOrder

	var orders []*match.EncryptedAuctionOrder
	// Add a bunch of orders to the order list
	var newEncrypted *match.EncryptedAuctionOrder
	for i := 0; i < testNumOrders; i++ {
		// We increment the nonce so orders are unique
		baseOrder.Nonce = incrementNonce(baseOrder.Nonce)
		if newEncrypted, err = (&baseOrder).TurnIntoEncryptedOrder(testStandardAuctionTime); err != nil {
			t.Errorf("Error turning base order into timelock encrypted order: %s", err)
		}
		orders = append(orders, newEncrypted)
	}

	t.Logf("%s: Placing all orders", time.Now())

	// Place a bunch of orders
	for _, order := range orders {
		// We can do this in sequence because it's going to start a goroutine anyways
		s.PlacePuzzledOrder(order)
		// Okay now we'll wait a little bit to make sure we see this memory issue in action
	}
	t.Logf("%s: Placed all orders", time.Now())

	t.Logf("%s: Ending auction", time.Now())
	// Wait for the auction to finish being batched
	var batchRes *match.AuctionBatch
	if batchRes, err = s.EndAuctionWithID(&testEncryptedOrder.IntendedPair, testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error ending auction with ID for TestUltraLightPlacePuzzledOrder: %s", err)
		return
	}
	t.Logf("%s: Ended auction", time.Now())

	t.Logf("%s: Matching orders", time.Now())
	if err = s.PlaceBatch(batchRes); err != nil {
		t.Errorf("Error placing and matching batch: %s", err)
		return
	}
	t.Logf("%s: Matched orders", time.Now())

	return
}

func TestPlaceNilOrderUltraLight(t *testing.T) {
	var err error

	var s *OpencxAuctionServer
	if s, err = initTestServer(); err != nil {
		t.Errorf("Error initializing test server for TestCreateServer: %s", err)
	}

	if err = s.PlacePuzzledOrder(nil); err == nil {
		t.Errorf("Placing a nil order succeeded! Should not be able to place a nil order!")
	}

	return
}
