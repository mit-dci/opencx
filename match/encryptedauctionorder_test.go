package match

import "testing"

func solveVariableRC5AuctionOrder(howMany uint64, timeToSolve uint64, t *testing.T) {

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

// TestConcurrentSolvesN10_T1000000 takes about 7.2 seconds on an i7 8700k
func TestConcurrentSolvesN10_T1000000(t *testing.T) {
	solveVariableRC5AuctionOrder(uint64(10), uint64(1000000), t)
	return
}
