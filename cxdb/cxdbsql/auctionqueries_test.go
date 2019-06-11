package cxdbsql

import (
	"bytes"
	"testing"

	"github.com/mit-dci/lit/coinparam"
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
	litereg, _       = match.AssetFromCoinParam(&coinparam.LiteRegNetParams)
	btcreg, _        = match.AssetFromCoinParam(&coinparam.RegressionNetParams)
	testAuctionOrder = &match.AuctionOrder{
		Pubkey:     [...]byte{0x02, 0xe7, 0xb7, 0xcf, 0xcf, 0x42, 0x2f, 0xdb, 0x68, 0x2c, 0x85, 0x02, 0xbf, 0x2e, 0xef, 0x9e, 0x2d, 0x87, 0x67, 0xf6, 0x14, 0x67, 0x41, 0x53, 0x4f, 0x37, 0x94, 0xe1, 0x40, 0xcc, 0xf9, 0xde, 0xb3},
		Nonce:      [2]byte{0x00, 0x00},
		AuctionID:  [32]byte{0xde, 0xad, 0xbe, 0xef},
		AmountWant: 100000,
		AmountHave: 10000,
		Side:       "buy",
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		Signature: []byte{0x1b, 0xd6, 0x0f, 0xd3, 0xec, 0x5b, 0x73, 0xad, 0xa9, 0x8a, 0x92, 0x79, 0x82, 0x0f, 0x8e, 0xab, 0xf8, 0x8f, 0x47, 0x6e, 0xc3, 0x15, 0x33, 0x72, 0xd9, 0x90, 0x51, 0x41, 0xfd, 0x0a, 0xa1, 0xa2, 0x4a, 0x73, 0x75, 0x4c, 0xa5, 0x28, 0x4a, 0xc2, 0xed, 0x5a, 0xe9, 0x33, 0x22, 0xf4, 0x41, 0x1f, 0x9d, 0xd1, 0x78, 0xb9, 0x17, 0xd4, 0xe9, 0x72, 0x51, 0x7f, 0x5b, 0xd7, 0xe5, 0x12, 0xe7, 0x69, 0xb0},
	}
	testEncryptedOrder, _ = testAuctionOrder.TurnIntoEncryptedOrder(testStandardAuctionTime)
	testEncryptedBytes, _ = testEncryptedOrder.Serialize()
	// examplePubkeyOne =
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
	var book map[float64][]*match.OrderIDPair
	if book, err = dbConn.ViewAuctionOrderBook(&testAuctionOrder.TradingPair, testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error vewing auction puzzle book, should not error: %s", err)
		return
	}

	if match.NumberOfOrders(book) != 0 {
		t.Errorf("Length of returned orderbook is %d, should be 0", len(book))
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

// auctionSequenceOne should
var (
	equalOrderLtcBtc = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "buy",
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 1000,
		AmountHave: 1000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x00, 0x01},
		Signature: []byte{0x22, 0x32, 0xff, 0xde, 0xad, 0xb1, 0x1e},
	}
	equalOrderCounterparty = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "sell",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 1000,
		AmountHave: 1000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x00, 0x31},
		Signature: []byte{0x23, 0x34, 0xf6, 0xd7, 0xa8, 0xb9, 0x19},
	}
	testingClearingPrice = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "sell",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 1000,
		AmountHave: 1000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x00, 0x31},
		Signature: []byte{0x23, 0x34, 0xf6, 0xd7, 0xa8, 0xb9, 0x19},
	}
	/*
		This order is a "buy" order, so they want assetWant and have assetHave.
		They have 1000 LTC, want 2000 BTC
		This would be 2.0 BTC/LTC, but considered a "sell" order for the pair BTC/LTC
		This would be 0.5 BTC/LTC, but considered a "buy" order for the pair LTC/BTC
	*/
	testingHalfBuy = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "buy",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 1000,
		AmountHave: 2000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x05, 0x11},
		Signature: []byte{0x23, 0x44, 0xf6, 0xd7, 0x58, 0xb9, 0x19},
	}
	/*
		This order is a "sell" order, so they want assetHave and have assetWant.
		They have 2000 BTC, want 1000 LTC
		This would be 2.0 BTC/LTC, but considered a "buy" order for the pair BTC/LTC
		This would be 0.5 BTC/LTC, but considered a "sell" order for the pair LTC/BTC
	*/
	testingHalfSell = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "sell",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 1000,
		AmountHave: 2000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x06, 0x17},
		Signature: []byte{0x23, 0x44, 0xf7, 0xd7, 0x58, 0xb9, 0x19},
	}

	testingDoubleSell = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "sell",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 2000,
		AmountHave: 1000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x10, 0x31},
		Signature: []byte{0x11, 0x34, 0xf6, 0xd7, 0xa8, 0xb9, 0x11},
	}
	testingDoubleBuy = &match.AuctionOrder{
		Pubkey: [33]byte{},
		Side:   "buy",
		// If we switch litereg and btcreg here then the db will complain about tables
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
		AmountWant: 2000,
		AmountHave: 1000,
		// lets see how the engine deals with this as well
		AuctionID: [32]byte{0x00, 0x00},
		Nonce:     [...]byte{0x10, 0x32},
		Signature: []byte{0x12, 0x24, 0xf6, 0xd7, 0xa8, 0xb9, 0x11},
	}
)

func TestClearingMatchingSimple(t *testing.T) {
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

	if err = dbConn.PlaceAuctionOrder(equalOrderLtcBtc); err != nil {
		t.Errorf("Error placing auction order equalOrderLtcBtc, should not error: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionOrder(equalOrderCounterparty); err != nil {
		t.Errorf("Error placing auction order equalOrderCounterparty, should not error: %s", err)
		return
	}

	var book map[float64][]*match.OrderIDPair
	if book, err = dbConn.ViewAuctionOrderBook(&equalOrderLtcBtc.TradingPair, equalOrderLtcBtc.AuctionID); err != nil {
		t.Errorf("There should not be an error matching the view auction orderbook: %s", err)
		return
	}

	expectedOrders := uint64(2)
	if match.NumberOfOrders(book) != expectedOrders {
		t.Errorf("Length of returned orderbook is %d, should be %d", len(book), expectedOrders)
		return
	}

	// The intended behavior is for these two to be matched since they have equal order sizes, should be the same trading pair,
	// and have equal auction IDs.
	var height uint64
	if height, err = dbConn.MatchAuction(equalOrderLtcBtc.AuctionID); err != nil {
		t.Errorf("The matching for the simple clearing test should not error: %s", err)
		return
	}

	// TODO: If height is removed, this height stuff is absolutely not necessary, change this test if refactoring auction matching
	t.Logf("Matched auction at height %d", height)

	// Since they are both matched, the orderbook should be empty... Or should it? Since this is an auction, should we just
	// create execution reports? The auctions are themselves points in time essentially, and having them immutable may be nice.
	// TODO: Define lots of behavior
	if book, err = dbConn.ViewAuctionOrderBook(&equalOrderLtcBtc.TradingPair, equalOrderLtcBtc.AuctionID); err != nil {
		t.Errorf("There should not be an error matching the view auction orderbook: %s", err)
		return
	}

	// okay so for now there should be nothing in here.
	if match.NumberOfOrders(book) != 0 {
		t.Errorf("Length of returned orderbook is %d, should be 0", len(book))
		return
	}

}

// TestClearingDoubleMatch is a test that places two orders that should both be matched at their midpoint clearing price.
func TestClearingDoubleMatch(t *testing.T) {
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

	if err = dbConn.PlaceAuctionOrder(testingHalfBuy); err != nil {
		t.Errorf("Error placing auction order testingHalfBuy, should not error: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionOrder(testingDoubleSell); err != nil {
		t.Errorf("Error placing auction order testingDoubleSell, should not error: %s", err)
		return
	}

	var book map[float64][]*match.OrderIDPair
	if book, err = dbConn.ViewAuctionOrderBook(&testingHalfSell.TradingPair, testingHalfSell.AuctionID); err != nil {
		t.Errorf("There should not be an error matching the view auction orderbook: %s", err)
		return
	}

	// There will be two orders in the orderbook now
	expectedOrders := uint64(2)
	if match.NumberOfOrders(book) != expectedOrders {
		t.Errorf("Length of returned orderbook is %d, should be %d", len(book), expectedOrders)
		return
	}

	// The intended behavior is for these two to be matched since they have equal order sizes, should be the same trading pair,
	// and have equal auction IDs.
	var height uint64
	if height, err = dbConn.MatchAuction(testingHalfSell.AuctionID); err != nil {
		t.Errorf("The matching for the simple clearing test should not error: %s", err)
		return
	}

	// TODO: If height is removed, this height stuff is absolutely not necessary, change this test if refactoring auction matching
	t.Logf("Matched auction at height %d", height)

	// Since they are both matched, the orderbook should be empty... Or should it? Since this is an auction, should we just
	// create execution reports? The auctions are themselves points in time essentially, and having them immutable may be nice.
	// TODO: Define lots of behavior
	if book, err = dbConn.ViewAuctionOrderBook(&testingHalfSell.TradingPair, testingHalfSell.AuctionID); err != nil {
		t.Errorf("There should not be an error matching the view auction orderbook: %s", err)
		return
	}

	// There will be 0 now that it's been matched
	expectedOrders = uint64(0)
	if match.NumberOfOrders(book) != expectedOrders {
		t.Errorf("Length of returned orderbook is %d, should be %d", len(book), expectedOrders)
		return
	}

}
