package match

import "testing"

func TestIsBuySide(t *testing.T) {

	buyAuction := &AuctionOrder{
		Side: "buy",
	}

	var res bool
	if res = buyAuction.IsBuySide(); !res {
		t.Errorf("Buy auction should have returned true, instead returned %t", res)
	}

	sellAuction := &AuctionOrder{
		Side: "sell",
	}

	if res = sellAuction.IsBuySide(); res {
		t.Errorf("Sell auction should have returned false, instead returned %t", res)
	}

	idkAuction := &AuctionOrder{
		Side: "idk",
	}

	if res = idkAuction.IsBuySide(); res {
		t.Errorf("Nonsense auction should have returned false, instead returned %t", res)
	}

	return
}

func TestIsSellSide(t *testing.T) {

	sellAuction := &AuctionOrder{
		Side: "sell",
	}

	var res bool
	if res = sellAuction.IsSellSide(); !res {
		t.Errorf("Sell auction should have returned true, instead returned %t", res)
	}

	buyAuction := &AuctionOrder{
		Side: "buy",
	}

	if res = buyAuction.IsSellSide(); res {
		t.Errorf("Buy auction should have returned false, instead returned %t", res)
	}

	idkAuction := &AuctionOrder{
		Side: "idk",
	}

	if res = idkAuction.IsSellSide(); res {
		t.Errorf("Nonsense auction should have returned false, instead returned %t", res)
	}

	return
}

func solveVariableRC5AuctionOrder(howMany uint64, timeToSolve uint64, t *testing.T) {

	orderPair := Pair{
		AssetWant: BTC,
		AssetHave: VTC,
	}

	// First create the order that will be puzzled and solved
	origOrder := &AuctionOrder{
		Side:        "buy",
		TradingPair: orderPair,
		AmountHave:  100000000,
		AmountWant:  100000000,
		// Just some bytes cause why not
		Nonce:          [2]byte{0xff, 0x12},
		OrderbookPrice: 1.00000000,
	}

	var encOrder *EncryptedAuctionOrder
	var err error
	if encOrder, err = origOrder.TurnIntoEncryptedOrder(timeToSolve); err != nil {
		t.Errorf("Error turning original test order into encrypted order. Test cannot proceed")
		return
	}

	puzzleResChan := make(chan *OrderPuzzleResult, howMany)
	for i := uint64(0); i < howMany; i++ {
		go SolveRC5AuctionOrderAsync(encOrder, puzzleResChan)
	}
	for i := uint64(0); i < howMany; i++ {
		var res *OrderPuzzleResult
		res = <-puzzleResChan
		if res.Err != nil {
			t.Errorf("Solving order puzzle returned an error: %s", res.Err)
			return
		}
	}

	return
}

// This should be super quick. Takes 0.1 seconds on an i7 8700k, most of the time is probably
// spent creating the test to solve.
func TestConcurrentSolvesN10_T10000(t *testing.T) {
	solveVariableRC5AuctionOrder(uint64(10), uint64(10000), t)
	return
}

// This should be less quick but still quick. Takes about 0.7 seconds on an i7 8700k
func TestConcurrentSolvesN10_T100000(t *testing.T) {
	solveVariableRC5AuctionOrder(uint64(10), uint64(100000), t)
	return
}

// TestConcurrentSolvesN10_T10000000 takes about 72 seconds on an i7 8700k
func TestConcurrentSolvesN10_T10000000(t *testing.T) {
	solveVariableRC5AuctionOrder(uint64(10), uint64(10000000), t)
	return
}
