package cxdbsql

import (
	"bytes"
	"testing"

	"github.com/mit-dci/opencx/match"
)

const (
	testStandardAuctionTime = 100
	// normal db user stuff
	testingUser = "testopencx"
	testingPass = "testpass"
	// root user stuff -- should be default
	rootUser = "root"
	rootPass = ""
	// test string to put before stuff in the config
	testString = "testopencxdb_"
)

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
	testEncryptedBytes, _ = testEncryptedOrder.Serialize()
)

// TestPlaceAuctionPuzzleGoodParams should succeed with the correct coin params.
func TestPlaceAuctionPuzzleGoodParams(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Skipf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("Error placing auction puzzle, should not error: %s", err)
		return
	}

	return
}

// TestPlaceAuctionPuzzleBadParams should succeed even with bad coin params.
func TestPlaceAuctionPuzzleBadParams(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("There was no error placing auction puzzle, should not error even w bad params: %s", err)
		return
	}

	return
}

// TestViewAuctionPuzzlebookEmpty tests that an empty orderbook doesn't error or return anything
func TestViewAuctionPuzzlebookEmpty(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	// Starting from an empty book, we shouldn't see anything in this auction id.
	var returnedOrders []*match.EncryptedAuctionOrder
	if returnedOrders, err = dbConn.ViewAuctionPuzzleBook(testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error vewing auction puzzle book, should not error: %s", err)
		return
	}

	if len(returnedOrders) != 0 {
		t.Errorf("Length of returned orders is %d, should be 0", len(returnedOrders))
		return
	}

	return
}

// TestViewAuctionOrderbookEmpty tests that an empty orderbook doesn't error or return anything
func TestViewAuctionOrderbookEmpty(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	// Starting from an empty book, we shouldn't see anything in this auction id.
	var retBuyOrders []*match.AuctionOrder
	var retSellOrders []*match.AuctionOrder
	if retBuyOrders, retSellOrders, err = dbConn.ViewAuctionOrderBook(&testAuctionOrder.TradingPair, testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error vewing auction puzzle book, should not error: %s", err)
		return
	}

	if len(retBuyOrders) != 0 {
		t.Errorf("Length of returned buy orders is %d, should be 0", len(retBuyOrders))
		return
	}

	if len(retSellOrders) != 0 {
		t.Errorf("Length of returned sell orders is %d, should be 0", len(retSellOrders))
		return
	}

	return
}

// TestPlaceAuctionPuzzlebookChanges should succeed with the correct coin params.
func TestPlaceAuctionPuzzlebookChanges(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("Error placing auction puzzle, should not error: %s", err)
		return
	}

	// Starting from an empty book, we should see this order added.
	var returnedOrders []*match.EncryptedAuctionOrder
	if returnedOrders, err = dbConn.ViewAuctionPuzzleBook(testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error vewing auction puzzle book, should not error: %s", err)
		return
	}

	if len(returnedOrders) != 1 {
		t.Errorf("Length of returned orders is %d, should be 1", len(returnedOrders))
		return
	}

	var retBytes []byte
	if retBytes, err = returnedOrders[0].Serialize(); err != nil {
		t.Errorf("Error serializing first returned order, should not error: %s", err)
		return
	}

	if bytes.Compare(retBytes, testEncryptedBytes) != 0 {
		t.Errorf("The serialized returned order from orderbook was not the same as the input, should be equal")
		return
	}

	return
}
