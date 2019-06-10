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

	return
}
