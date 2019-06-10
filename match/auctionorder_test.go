package match

import (
	"bytes"
	"testing"
)

var (
	orderPair = Pair{
		AssetWant: BTC,
		AssetHave: VTC,
	}

	origOrderID = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	origOrder   = &AuctionOrder{
		Side:        "buy",
		TradingPair: orderPair,
		AmountHave:  100000000,
		AmountWant:  100000000,
		// Just some bytes cause why not
		Nonce: [2]byte{0xff, 0x12},
	}

	origOrderCounterID = []byte{0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}
	origOrderCounter   = &AuctionOrder{
		Side:        "sell",
		TradingPair: orderPair,
		AmountHave:  100000000,
		AmountWant:  100000000,
		// Just some bytes cause why not
		Nonce: [2]byte{0xff, 0x12},
	}

	origOrderFullExec = &OrderExecution{
		OrderID: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		Debited: Entry{
			Amount: 100000000,
			Asset:  BTC,
		},
		Credited: Entry{
			Amount: 100000000,
			Asset:  VTC,
		},
		// these are just some random numbers because they should not matter since the order is filled
		NewAmountWant: 23892323,
		NewAmountHave: 37348722,
		Filled:        true,
	}

	origOrderDoubleExec = &OrderExecution{
		OrderID: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		Debited: Entry{
			Amount: 200000000,
			Asset:  BTC,
		},
		Credited: Entry{
			Amount: 100000000,
			Asset:  VTC,
		},
		// these are just some random numbers because they should not matter since the order is filled
		NewAmountWant: 53872666,
		NewAmountHave: 47666772,
		Filled:        true,
	}
)

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

// Test some easy execution generation
func TestGenerateEasyExecutionFromPrice(t *testing.T) {
	var err error

	// this should fill the order completely. this is the trivial case.
	var resExec OrderExecution
	var fillRemainder uint64
	if resExec, fillRemainder, err = origOrder.GenerateExecutionFromPrice(origOrderID, float64(1), 100000000); err != nil {
		t.Errorf("Error generating execution from price, should not error: %s", err)
		return
	}
	// while they shouldn't be equal, the non Amount fields should be.
	if resExec.Filled != origOrderFullExec.Filled {
		t.Errorf("Both executions should be filled, but the result's filled variable is %t", resExec.Filled)
		return
	}
	if !bytes.Equal(resExec.OrderID, origOrderFullExec.OrderID) {
		t.Errorf("Order IDs should be equal for both executions. The result should be %x but was %x", origOrderFullExec.OrderID, resExec.OrderID)
		return
	}
	if resExec.Credited != origOrderFullExec.Credited {
		t.Errorf("Executions should have the same amount and asset credited. The result should be %s but was %s", &origOrderFullExec.Credited, &resExec.Credited)
		return
	}
	if resExec.Debited != origOrderFullExec.Debited {
		t.Errorf("Executions should have the same amount and asset debited. The result should be %s but was %s", &origOrderFullExec.Debited, &resExec.Debited)
		return
	}
	if resExec.NewAmountHave != 0 {
		t.Errorf("A filled order should have no 'NewAmountHave' left. It has %d instead", resExec.NewAmountHave)
		return
	}
	if fillRemainder != 0 {
		t.Errorf("The remainder from this order fill should be 0 since it should be exact. Was %d instead", fillRemainder)
		return
	}

	return
}

// Test execution generation for a price that will fill for "half price", aka give the orderID's user twice the money
func TestGenerateDoubleFill(t *testing.T) {
	var err error

	// this should fill the order completely. this is the trivial case.
	var resExec OrderExecution
	if resExec, err = origOrder.GenerateOrderFill(origOrderID, float64(2)); err != nil {
		t.Errorf("Error generating execution from price, should not error: %s", err)
		return
	}
	// while they shouldn't be equal, the non Amount fields should be.
	if resExec.Filled != origOrderDoubleExec.Filled {
		t.Errorf("Both executions should be filled, but the result's filled variable is %t", resExec.Filled)
		return
	}
	if !bytes.Equal(resExec.OrderID, origOrderDoubleExec.OrderID) {
		t.Errorf("Order IDs should be equal for both executions. The result should be %x but was %x", origOrderDoubleExec.OrderID, resExec.OrderID)
		return
	}
	if resExec.Credited != origOrderDoubleExec.Credited {
		t.Errorf("Executions should have the same amount and asset credited. The result should be %s but was %s", &origOrderDoubleExec.Credited, &resExec.Credited)
		return
	}
	if resExec.Debited != origOrderDoubleExec.Debited {
		t.Errorf("Executions should have the same amount and asset debited. The result should be %s but was %s", &origOrderDoubleExec.Debited, &resExec.Debited)
		return
	}
	if resExec.NewAmountHave != 0 {
		t.Errorf("A filled order should have no 'NewAmountHave' left. It has %d instead", resExec.NewAmountHave)
		return
	}

	return
}

// Test some easy fill generation
func TestGenerateEasyFillFromPrice(t *testing.T) {
	var err error

	// this should fill the order completely. this is the trivial case.
	var resExec OrderExecution
	if resExec, err = origOrder.GenerateOrderFill(origOrderID, float64(1)); err != nil {
		t.Errorf("Error generating execution from price, should not error: %s", err)
		return
	}
	// while they shouldn't be equal, the non Amount fields should be.
	if resExec.Filled != origOrderFullExec.Filled {
		t.Errorf("Both executions should be filled, but the result's filled variable is %t", resExec.Filled)
		return
	}
	if !bytes.Equal(resExec.OrderID, origOrderFullExec.OrderID) {
		t.Errorf("Order IDs should be equal for both executions. The result should be %x but was %x", origOrderFullExec.OrderID, resExec.OrderID)
		return
	}
	if resExec.Credited != origOrderFullExec.Credited {
		t.Errorf("Executions should have the same amount and asset credited. The result should be %s but was %s", &origOrderFullExec.Credited, &resExec.Credited)
		return
	}
	if resExec.Debited != origOrderFullExec.Debited {
		t.Errorf("Executions should have the same amount and asset debited. The result should be %s but was %s", &origOrderFullExec.Debited, &resExec.Debited)
		return
	}
	if resExec.NewAmountHave != 0 {
		t.Errorf("A filled order should have no 'NewAmountHave' left. It has %d instead", resExec.NewAmountHave)
		return
	}

	return
}

// Test some fill generation based on a bad side that should error out
func TestGenerateBadSideFill(t *testing.T) {
	var err error

	// Create a new order that looks like origOrder
	badOrder := new(AuctionOrder)
	*badOrder = *origOrder
	badOrder.Side = "bad"

	// this should just error
	var resExec OrderExecution
	if resExec, err = badOrder.GenerateOrderFill(origOrderID, float64(1)); err == nil {
		t.Errorf("There was no error trying to generate an order fill for an order with a bad side")
		return
	}

	emptyExec := &OrderExecution{}
	if !(&resExec).Equal(emptyExec) {
		t.Errorf("GenerateOrderFill created part of an execution on failure, this should not happen")
		return
	}

	return
}

// Test some fill generation based on a zero price that should error out
func TestGenerateBadPriceFill(t *testing.T) {
	var err error

	// Create a new order that looks like origOrder
	badOrder := new(AuctionOrder)
	*badOrder = *origOrder

	// this should just error
	var resExec OrderExecution
	if resExec, err = badOrder.GenerateOrderFill(origOrderID, float64(0)); err == nil {
		t.Errorf("There was no error trying to generate an order fill for a price of zero")
		return
	}

	emptyExec := &OrderExecution{}
	if !(&resExec).Equal(emptyExec) {
		t.Errorf("GenerateOrderFill created part of an execution on failure, this should not happen")
		return
	}

	return
}

// Test some fill generation on an orderbook price of zero
func TestGenerateEasyPriceFillAmounts(t *testing.T) {
	var err error

	// Create a new order that looks like origOrder
	zeroPriceOrder := new(AuctionOrder)
	*zeroPriceOrder = *origOrder

	// this should just error
	var resExec OrderExecution
	if resExec, err = zeroPriceOrder.GenerateOrderFill(origOrderID, float64(1)); err != nil {
		t.Errorf("Error generating execution from price, should not error: %s", err)
		return
	}

	// while they shouldn't be equal, the non Amount fields should be.
	if resExec.Filled != origOrderFullExec.Filled {
		t.Errorf("Both executions should be filled, but the result's filled variable is %t", resExec.Filled)
		return
	}
	if !bytes.Equal(resExec.OrderID, origOrderFullExec.OrderID) {
		t.Errorf("Order IDs should be equal for both executions. The result should be %x but was %x", origOrderFullExec.OrderID, resExec.OrderID)
		return
	}
	if resExec.Credited != origOrderFullExec.Credited {
		t.Errorf("Executions should have the same amount and asset credited. The result should be %s but was %s", &origOrderFullExec.Credited, &resExec.Credited)
		return
	}
	if resExec.Debited != origOrderFullExec.Debited {
		t.Errorf("Executions should have the same amount and asset debited. The result should be %s but was %s", &origOrderFullExec.Debited, &resExec.Debited)
		return
	}
	if resExec.NewAmountHave != 0 {
		t.Errorf("A filled order should have no 'NewAmountHave' left. It has %d instead", resExec.NewAmountHave)
		return
	}

	return
}

// Test a very simple price (1) and make sure that the price calculation is the same for both buy and sell
func TestSimplePriceValidBuy(t *testing.T) {
	var err error

	var retPriceOne float64
	if retPriceOne, err = origOrder.Price(); err != nil {
		t.Errorf("Calculating price for origOrder should not have failed, here's the err: %s", err)
		return
	}

	expectedPrice := 1.0
	if retPriceOne != expectedPrice {
		t.Errorf("Price for origOrder should have been %f but was %f", expectedPrice, retPriceOne)
		return
	}

	var retPriceOneCounter float64
	if retPriceOneCounter, err = origOrderCounter.Price(); err != nil {
		t.Errorf("Calculating price for origOrderCounter should not have failed, here's the err: %s", err)
		return
	}

	expectedPriceCounter := 1.0
	if retPriceOneCounter != expectedPriceCounter {
		t.Errorf("Price for origOrderCounter should have been %f but was %f", expectedPriceCounter, retPriceOneCounter)
		return
	}

	if retPriceOneCounter != retPriceOne {
		t.Errorf("The price for retPriceOne, which was %f, should have been the same as retPriceOneCounter, which was %f", retPriceOne, retPriceOneCounter)
		return
	}

}

// In Binance, BTC/USD price is always in the ratio USD/BTC.
// Higher is better (for sells) and lower is better (for buys).
// This is completely backwards, since if I see the price 9000 on the pair BTC/USD,
// then I should treat the price as the ration BTC / USD, not the other way around.

// "buy" is categorized as having usd, and wanting btc.
// "sell" is categorized as having btc, and wanting usd.
// So price is always in the ratio assetHave/assetWant.
// Ideally the orderbook will show both prices assetWant/assetHave and assetHave/assetWant.
// But for our purposes, since we've modeled it as a ratio we're sticking with that.
var (
	// Pair:
	//     Want: BTC
	//     Have: LTC
	// This order is meant to be a test of the representation of price
	// Since the user is a buyer (buyer of BTC in the BTC/LTC pair), they have LTC and want BTC
	// So if the price is assetWant / assetHave, then this will be a price of 2 BTC/LTC. Simple enough.
	// That price is formatted so well you could do dimensional analysis on it. Unlike on Binance.
	priceTwoBuy = &AuctionOrder{
		Side:        "buy",
		TradingPair: orderPair,
		AmountWant:  200000000, // BTC - This user wants this asset
		AmountHave:  100000000, // LTC - This user has this asset
		// Just some bytes cause why not
		Nonce: [2]byte{0xff, 0x12},
	}
	// So if the price is assetWant / assetHave (To get the ratio BTC/LTC), then this will be a price of 2 BTC/LTC.
	priceTwoSell = &AuctionOrder{
		Side:        "sell",
		TradingPair: orderPair,
		AmountWant:  200000000, // BTC - This user has this asset
		AmountHave:  100000000, // LTC - This user wants this asset
		// Just some bytes cause why not
		Nonce: [2]byte{0xf1, 0x23},
	}
	// This should error on price because AmountWant = 0
	priceErrorWant = &AuctionOrder{
		Side:        "sell",
		TradingPair: orderPair,
		AmountWant:  0,         // BTC - This user has this asset
		AmountHave:  100000000, // LTC - This user wants this asset
		// Just some bytes cause why not
		Nonce: [2]byte{0xf1, 0x23},
	}
	// This should error on price because AmountHave = 0
	priceErrorHave = &AuctionOrder{
		Side:        "buy",
		TradingPair: orderPair,
		AmountWant:  200000000, // BTC - This user has this asset
		AmountHave:  0,         // LTC - This user wants this asset
		// Just some bytes cause why not
		Nonce: [2]byte{0xf1, 0x23},
	}
	// This should error on price because both are = 0
	priceErrorBoth = &AuctionOrder{
		Side:        "buy",
		TradingPair: orderPair,
		AmountWant:  0, // BTC - This user has this asset
		AmountHave:  0, // LTC - This user wants this asset
		// Just some bytes cause why not
		Nonce: [2]byte{0xf1, 0x23},
	}
)

// validPriceTest runs a test to make sure the order has price expectedPrice
func validPriceTest(order *AuctionOrder, expectedPrice float64, t *testing.T) {
	var err error

	var origPrice float64
	if origPrice, err = order.Price(); err != nil {
		t.Errorf("Error getting price for order: %s", err)
		return
	}

	if origPrice != expectedPrice {
		t.Errorf("Test failed: price should have been %f but was %f", expectedPrice, origPrice)
		return
	}

	return
}

// validPriceTest runs a test to make sure the order has an error for calculating its price
func errorPriceTest(order *AuctionOrder, t *testing.T) {
	var err error

	var origPrice float64
	if origPrice, err = order.Price(); err == nil {
		t.Errorf("There was no error while calculating price for order, instead a price of %f was returned", origPrice)
		return
	}

	return
}

func TestPriceOneEasy(t *testing.T) {
	validPriceTest(origOrder, float64(1), t)
	return
}

func TestPriceTwoBuy(t *testing.T) {
	validPriceTest(priceTwoBuy, float64(2), t)
	return
}

func TestPriceTwoSell(t *testing.T) {
	validPriceTest(priceTwoSell, float64(2), t)
	return
}

func TestErrorHave(t *testing.T) {
	errorPriceTest(priceErrorHave, t)
	return
}

func TestErrorWant(t *testing.T) {
	errorPriceTest(priceErrorWant, t)
	return
}

func TestErrorBoth(t *testing.T) {
	errorPriceTest(priceErrorBoth, t)
	return
}

// TODO add more tests for simple methods
