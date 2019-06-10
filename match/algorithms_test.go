package match

import (
	"fmt"
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"golang.org/x/crypto/sha3"
)

var (
	litereg, _ = AssetFromCoinParam(&coinparam.LiteRegNetParams)
	btcreg, _  = AssetFromCoinParam(&coinparam.RegressionNetParams)
	BTC_LTC    = &Pair{
		AssetWant: btcreg,
		AssetHave: litereg,
	}
	onePriceSell = &AuctionOrder{
		Side:        "sell",
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  1000,
	}
	onePriceBuy = &AuctionOrder{
		Side:        "buy",
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  1000,
	}
)

// generateLargeClearingBook puts a bunch of sell orders on the side that should be cleared, and a bunch of buy orders on the side that should be cleared
func generateLargeClearingBook(midpoint float64, radius uint64) (book map[float64][]*OrderIDPair, err error) {
	floatIncrement := midpoint / float64(radius)
	if floatIncrement <= float64(0) {
		err = fmt.Errorf("floatIncrement would not have been enough. Try again with different parameters")
		return
	}

	var orders []*AuctionOrder
	var thisOrder *AuctionOrder
	for i := uint64(0); i < 2*radius; i++ {
		thisOrder = &AuctionOrder{
			TradingPair: *BTC_LTC,
			AmountWant:  1000,
			AmountHave:  1000,
		}
		if i < radius {
			thisOrder.Side = "buy"
		} else {
			thisOrder.Side = "sell"
		}
		orders = append(orders, thisOrder)
	}

	if book, err = createBookFromOrders(orders); err != nil {
		err = fmt.Errorf("Error creating book from orders while generating large clearing book: %s", err)
		return
	}

	return
}

func createBookFromOrders(orders []*AuctionOrder) (book map[float64][]*OrderIDPair, err error) {
	book = make(map[float64][]*OrderIDPair)
	var pr float64
	for _, order := range orders {
		if pr, err = order.Price(); err != nil {
			err = fmt.Errorf("Error getting price from order while creating book from orders: %s", err)
			return
		}
		book[pr] = append(book[pr], &OrderIDPair{
			OrderID: sha3.Sum256(order.SerializeSignable()),
			Order:   order,
		})
	}
	return
}

func runLargeClearingBookTest(midpoint float64, orderRadius uint64, t *testing.T) {

	var err error

	var fakeNeutralBook map[float64][]*OrderIDPair
	if fakeNeutralBook, err = generateLargeClearingBook(midpoint, orderRadius); err != nil {
		t.Errorf("Error creating book from orders for test: %s", err)
		return
	}

	// Test execs at clearing price 1 (thats the price so yeah)
	var execs []*OrderExecution
	if execs, err = ClearingMatchingAlgorithm(fakeNeutralBook, midpoint); err != nil {
		t.Errorf("Error running clearing matching algorithm for test: %s", err)
		return
	}
	if uint64(len(execs)) != orderRadius {
		t.Errorf("There should have been %d executions, instead there are %d", orderRadius, len(execs))
		return
	}
	for _, exec := range execs {
		if !exec.Filled {
			t.Errorf("All orders should have been filled. There is an unfilled order.")
			return
		}
		if exec.NewAmountWant != 0 {
			t.Errorf("All orders should have zero NewAmountWant, this was %d", exec.NewAmountWant)
			return
		}
		if exec.NewAmountHave != 0 {
			t.Errorf("All orders should have zero NewAmountHave, this was %d", exec.NewAmountHave)
			return
		}
	}

	return
}

func TestClearingPriceSamePriceBook(t *testing.T) {
	var err error

	ordersToInsert := []*AuctionOrder{onePriceBuy, onePriceSell}

	var fakeNeutralBook map[float64][]*OrderIDPair
	if fakeNeutralBook, err = createBookFromOrders(ordersToInsert); err != nil {
		t.Errorf("Error creating book from orders for test: %s", err)
		return
	}

	// Test execs at clearing price 1 (thats the price so yeah)
	var execs []*OrderExecution
	if execs, err = ClearingMatchingAlgorithm(fakeNeutralBook, float64(1)); err != nil {
		t.Errorf("Error running clearing matching algorithm for test: %s", err)
		return
	}
	if len(execs) != 2 {
		t.Errorf("There should have been 2 executions, instead there are %d", len(execs))
		return
	}
	for _, exec := range execs {
		if !exec.Filled {
			t.Errorf("All orders should have been filled. There is an unfilled order.")
			return
		}
		if exec.NewAmountWant != 0 {
			t.Errorf("All orders should have zero NewAmountWant, this was %d", exec.NewAmountWant)
			return
		}
		if exec.NewAmountHave != 0 {
			t.Errorf("All orders should have zero NewAmountHave, this was %d", exec.NewAmountHave)
			return
		}
		if exec.Credited.Amount != 1000 {
			t.Errorf("All orders should have credited amount = 1000, this was %d", exec.Credited.Amount)
			return
		}
		if exec.Debited.Amount != 1000 {
			t.Errorf("All orders should have debited amount = 1000, this was %d", exec.Debited.Amount)
			return
		}
	}

	return
}

func TestClearingPrice1000_25Clear(t *testing.T) {
	runLargeClearingBookTest(float64(25), 1000, t)
	return
}
