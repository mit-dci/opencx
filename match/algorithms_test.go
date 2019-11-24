package match

import (
	"fmt"
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"golang.org/x/crypto/sha3"
)

// Rationale behind clearing price matching:
// On a typical exchange, if, for example you have US dollars and are trying to buy BTC,
// then a higher price for your order will get you priority, but a lower price is always better
// for you, and you will always take a lower price if you can get it.
// The price is determined as USD/BTC, or have/want.
// In our model, if you have one asset and are trying to buy another asset then your price will
// be in want/have.
// If you are a buyer, this means a lower price will get you priority, but you will take any
// price that is higher.
// If you are a seller, the opposite is true.
// So in clearing matching, if I were to have an order that is priced at $5/btc, and I am a buyer
// then I will get priority over an order that is $4/btc because 1/5 < 1/4. If I receive a price
// of 1btc/4usd, or $4/btc, then I will still be satisfied. I will only have to give up 4 of my
// dollars.
// If I am a seller, and I put in an order priced at $3/btc, or 1btc/3usd, then I will get priority
// over any sell order > 1/3, since that means those orders want more usd for the same amount of btc.
// If I receive a price of 1btc/4usd, or $4/btc, then I will still be satisfied, since I will get
// more usd and give up the same amount of btc.
// This price of $4/btc, or 1btc/4usd, will be one of our trivial tests.
var (
	litereg, _ = AssetFromCoinParam(&coinparam.LiteRegNetParams)
	btcreg, _  = AssetFromCoinParam(&coinparam.RegressionNetParams)
	BTC_LTC    = &Pair{
		AssetWant: btcreg,
		AssetHave: litereg,
	}
	onePriceSell = &AuctionOrder{
		Side:        Sell,
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  1000,
	}
	onePriceBuy = &AuctionOrder{
		Side:        Buy,
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  1000,
	}
	trivialQuarterBuy = &AuctionOrder{
		Side:        Buy,
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  5000,
	}
	trivialQuarterSell = &AuctionOrder{
		Side:        Sell,
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  3000,
	}
)

// generateLargeClearingBook puts a bunch of sell orders on the side that should be cleared, and a bunch of buy orders on the side that should be cleared
func generateLargeClearingBook(midpoint float64, radius uint64) (book map[float64][]*AuctionOrderIDPair, err error) {
	floatIncrement := midpoint / float64(radius)
	if floatIncrement <= float64(0) {
		err = fmt.Errorf("floatIncrement would not have been enough. Try again with different parameters")
		return
	}

	var orders []*AuctionOrder
	var thisOrder *AuctionOrder
	for i := uint64(1); i < 2*radius; i++ {
		thisOrder = &AuctionOrder{
			TradingPair: *BTC_LTC,
			AmountWant:  uint64(float64(100000000) * float64(i) * floatIncrement),
			AmountHave:  100000000,
		}
		// Lower end of the price range for buy means it's more
		// competitive. The least competitive buy order still matches.
		if i < radius {
			thisOrder.Side = Buy
			// Higher end of the price range for sell means it's more
			// competitive. The least competitive sell order still
			// matches.
		} else {
			thisOrder.Side = Sell
		}
		orders = append(orders, thisOrder)
	}

	if book, err = createBookFromOrders(orders); err != nil {
		err = fmt.Errorf("Error creating book from orders while generating large clearing book: %s", err)
		return
	}

	return
}

func createBookFromOrders(orders []*AuctionOrder) (book map[float64][]*AuctionOrderIDPair, err error) {
	book = make(map[float64][]*AuctionOrderIDPair)
	var pr float64
	for _, order := range orders {
		if pr, err = order.Price(); err != nil {
			err = fmt.Errorf("Error getting price from order while creating book from orders: %s", err)
			return
		}
		book[pr] = append(book[pr], &AuctionOrderIDPair{
			OrderID: sha3.Sum256(order.SerializeSignable()),
			Order:   order,
		})
	}
	return
}

func runLargeClearingBookTest(midpoint float64, orderRadius uint64, t *testing.T) {
	var err error

	var fakeNeutralBook map[float64][]*AuctionOrderIDPair
	if fakeNeutralBook, err = generateLargeClearingBook(midpoint, orderRadius); err != nil {
		t.Errorf("Error creating book from orders for test: %s", err)
		return
	}

	// We get the number of orders here because we know all of them should execute.
	numOrders := NumberOfOrders(fakeNeutralBook)

	// Test execs at clearing price 1 (thats the price so yeah)
	var execs []*OrderExecution
	var setExecs []*SettlementExecution
	if execs, setExecs, err = MatchClearingAlgorithm(fakeNeutralBook); err != nil {
		t.Errorf("Error running clearing matching algorithm for test: %s", err)
		return
	}
	if uint64(len(setExecs)) != numOrders*2 {
		t.Errorf("There should have been %d settlement executions, instead there are %d", numOrders, len(setExecs))
		return
	}
	if uint64(len(execs)) != numOrders {
		t.Errorf("There should have been %d executions, instead there are %d", numOrders, len(execs))
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

	var fakeNeutralBook map[float64][]*AuctionOrderIDPair
	if fakeNeutralBook, err = createBookFromOrders(ordersToInsert); err != nil {
		t.Errorf("Error creating book from orders for test: %s", err)
		return
	}

	// Test execs at clearing price 1 (thats the price so yeah)
	var execs []*OrderExecution
	var setExecs []*SettlementExecution
	if execs, setExecs, err = MatchClearingAlgorithm(fakeNeutralBook); err != nil {
		t.Errorf("Error running clearing matching algorithm for test: %s", err)
		return
	}
	if len(execs) != 2 {
		t.Errorf("There should have been 2 order executions, instead there are %d", len(execs))
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

	for _, setExec := range setExecs {
		if setExec.Type == Credit && setExec.Amount != 1000 {
			t.Errorf("All orders should have credited amount = 1000, this was %d", setExec.Amount)
			return
		}
		if setExec.Type == Debit && setExec.Amount != 1000 {
			t.Errorf("All orders should have debited amount = 1000, this was %d", setExec.Amount)
			return
		}

	}

	return
}

func TestClearingTrivial(t *testing.T) {
	var err error

	ordersToInsert := []*AuctionOrder{trivialQuarterBuy, trivialQuarterSell}

	var fakeNeutralBook map[float64][]*AuctionOrderIDPair
	if fakeNeutralBook, err = createBookFromOrders(ordersToInsert); err != nil {
		t.Errorf("Error creating book from orders for test: %s", err)
		return
	}

	// Test execs at clearing price 1 (thats the price so yeah)
	var execs []*OrderExecution
	var setExecs []*SettlementExecution
	if execs, setExecs, err = MatchClearingAlgorithm(fakeNeutralBook); err != nil {
		t.Errorf("Error running clearing matching algorithm for test: %s", err)
		return
	}
	if len(setExecs) != 4 {
		t.Errorf("There should have been 4 settlement executions, instead there are %d", len(setExecs))
		return
	}
	if len(execs) != 2 {
		t.Errorf("There should have been 4 order executions, instead there are %d", len(execs))
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

func TestClearingPrice1000_250Clear(t *testing.T) {
	runLargeClearingBookTest(float64(250), 1000, t)
	return
}

func TestClearingPrice1000_25Clear(t *testing.T) {
	runLargeClearingBookTest(float64(25), 1000, t)
	return
}

func TestClearingPrice10000_1Clear(t *testing.T) {
	runLargeClearingBookTest(float64(1), 10000, t)
	return
}

// BenchmarkAlgorithmsClearingPrice benchmarks the clearing price
// matching algorithm
func BenchmarkAlgorithmsClearingPrice(b *testing.B) {
	var err error
	b.StopTimer()
	b.ResetTimer()

	midpointForClearingBook := float64(150)
	orderRadiusForBook := uint64(10000)

	var fakeNeutralBook map[float64][]*AuctionOrderIDPair
	if fakeNeutralBook, err = generateLargeClearingBook(midpointForClearingBook, orderRadiusForBook); err != nil {
		b.Fatalf("Error creating book from orders for test: %s", err)
		return
	}

	b.StartTimer()
	// Test execs at clearing price
	for i := 0; i < b.N; i++ {
		_, _, err = MatchClearingAlgorithm(fakeNeutralBook)
	}
	b.StopTimer()

	// We check errors at the end with a fixed input because we don't
	// really want to waste time error checking (even if it's fast)
	// during the testing, it's the same input and the regular unit
	// tests should be covering cases where this would break
	if err != nil {
		b.Fatalf("Error running clearing matching algorithm for test: %s", err)
		return
	}

	return
}
