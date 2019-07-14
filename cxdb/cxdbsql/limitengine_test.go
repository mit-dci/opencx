package cxdbsql

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

var (
	testLimitOrder = &match.LimitOrder{
		Pubkey:     [...]byte{0x02, 0xe7, 0xb7, 0xcf, 0xcf, 0x42, 0x2f, 0xdb, 0x68, 0x2c, 0x85, 0x02, 0xbf, 0x2e, 0xef, 0x9e, 0x2d, 0x87, 0x67, 0xf6, 0x14, 0x67, 0x41, 0x53, 0x4f, 0x37, 0x94, 0xe1, 0x40, 0xcc, 0xf9, 0xde, 0xb3},
		AmountWant: 100000,
		AmountHave: 10000,
		Side:       match.Buy,
		TradingPair: match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		},
	}
	// examplePubkeyOne =
)

func TestCreateLimitEngineAllParams(t *testing.T) {
	var err error

	var killFunc func(t *testing.T)
	if killFunc, err = createUserAndDatabase(); err != nil {
		t.Skipf("Error creating user and database, skipping: %s", err)
	}

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(constCoinParams()); err != nil {
		t.Errorf("Error creating asset pairs from coin list: %s", err)
	}

	for _, pair := range pairList {
		if _, err = CreateLimitEngineWithConf(pair, testConfig()); err != nil {
			t.Errorf("Error creating auction engine for pair: %s", err)
		}
	}

	killFunc(t)
}

func TestPlaceSingleLimitOrder(t *testing.T) {
	var err error

	var killFunc func(t *testing.T)
	if killFunc, err = createUserAndDatabase(); err != nil {
		t.Skipf("Error creating user and database, skipping: %s", err)
	}

	var engine match.LimitEngine
	if engine, err = CreateLimitEngineWithConf(&testLimitOrder.TradingPair, testConfig()); err != nil {
		t.Errorf("Error creating limit engine for pair: %s", err)
	}

	if _, err = engine.PlaceLimitOrder(testLimitOrder); err != nil {
		t.Errorf("Error placing limit order: %s", err)
	}

	killFunc(t)
}

func BenchmarkAllLimit(b *testing.B) {

	for _, howMany := range []uint64{1, 10, 100, 1000, 10000, 100000} {
		PlaceNLimitOrders(howMany, b)
		MatchNLimitOrders(howMany, b)
		PlaceMatchNLimitOrders(howMany, b)
	}

}

func PlaceNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	b.Run(fmt.Sprintf("Place%dLimitOrders", howMany), func(b *testing.B) {

		// Stop the timer because we're doing initialization stuff
		b.StopTimer()

		// We do all of this within the loop because otherwise the database unfortunately persists and we get duplicate key errors
		var killFunc func(b *testing.B)
		if killFunc, err = createUserAndDatabaseBench(); err != nil {
			b.Skipf("Error creating user and database, skipping: %s", err)
		}

		var engine match.LimitEngine
		if engine, err = CreateLimitEngineWithConf(&testLimitOrder.TradingPair, testConfig()); err != nil {
			b.Errorf("Error creating limit engine for pair: %s", err)
		}

		// Start it back up again, let's time this
		b.StartTimer()

		for _, order := range ordersToPlace {
			if _, err = engine.PlaceLimitOrder(order); err != nil {
				b.Errorf("Error placing limit order: %s", err)
			}
		}

		// We've got all that we need, let's stop it
		b.StopTimer()

		killFunc(b)
	})

}

func MatchNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	b.Run(fmt.Sprintf("Match%dLimitOrders", howMany), func(b *testing.B) {

		// Stop the timer because we're doing initialization stuff
		b.StopTimer()

		// We do all of this within the loop because otherwise the database unfortunately persists and we get duplicate key errors
		var killFunc func(b *testing.B)
		if killFunc, err = createUserAndDatabaseBench(); err != nil {
			b.Skipf("Error creating user and database, skipping: %s", err)
		}

		var engine match.LimitEngine
		if engine, err = CreateLimitEngineWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
			b.Errorf("Error creating limit engine for pair: %s", err)
		}

		for _, order := range ordersToPlace {
			if _, err = engine.PlaceLimitOrder(order); err != nil {
				b.Errorf("Error placing limit order: %s", err)
			}
		}

		// Start it back up again, let's time this
		b.StartTimer()

		if _, _, err = engine.MatchLimitOrders(); err != nil {
			b.Errorf("Error matching limit orders: %s", err)
		}

		// We've got all that we need, let's stop it
		b.StopTimer()

		killFunc(b)
	})

}

// PlaceMatchNLimitOrders is a bit different but more real, it places an order then runs matching. This should get an idea of overall throughput, maybe cause some errors
func PlaceMatchNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	b.Run(fmt.Sprintf("PlaceMatch%dLimitOrders", howMany), func(b *testing.B) {

		// Stop the timer because we're doing initialization stuff
		b.StopTimer()

		// We do all of this within the loop because otherwise the database unfortunately persists and we get duplicate key errors
		var killFunc func(b *testing.B)
		if killFunc, err = createUserAndDatabaseBench(); err != nil {
			b.Skipf("Error creating user and database, skipping: %s", err)
		}

		var engine match.LimitEngine
		if engine, err = CreateLimitEngineWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
			b.Errorf("Error creating limit engine for pair: %s", err)
		}

		// TODO: handle unhandled, erroring matching case where you have one order in, or there isn't a min / max price for buy/sell

		// Start it back up again, let's time this
		b.StartTimer()
		for _, order := range ordersToPlace {
			if _, err = engine.PlaceLimitOrder(order); err != nil {
				b.Errorf("Error placing limit order: %s", err)
			}
			if _, _, err = engine.MatchLimitOrders(); err != nil {
				b.Errorf("Error matching limit orders: %s", err)
			}
		}
		// We've got all that we need, let's stop it
		b.StopTimer()

		killFunc(b)
	})

}

// fuzzManyLimitOrders creates a bunch of orders that have some random amounts, to make the clearing price super random
// It creates a new private key for every order, signs it after constructing it, etc.
func fuzzManyLimitOrders(howMany uint64, pair match.Pair) (orders []*match.LimitOrder, err error) {
	// want this to be seeded so it's reproducible

	// 21 million is max amount
	maxAmount := int64(2100000000000000)
	r := rand.New(rand.NewSource(1801))
	orders = make([]*match.LimitOrder, howMany)
	for i := uint64(0); i < howMany; i++ {
		currOrder := new(match.LimitOrder)
		var ephemeralPrivKey *koblitz.PrivateKey
		if ephemeralPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			err = fmt.Errorf("Error generating new private key: %s", err)
			return
		}
		currOrder.AmountWant = uint64(r.Int63n(maxAmount) + 1)
		currOrder.AmountHave = uint64(r.Int63n(maxAmount) + 1)
		if r.Int63n(2) == 0 {
			currOrder.Side = match.Buy
		} else {
			currOrder.Side = match.Sell
		}
		currOrder.TradingPair = match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		}
		pubkeyBytes := ephemeralPrivKey.PubKey().SerializeCompressed()
		copy(currOrder.Pubkey[:], pubkeyBytes)

		// It's all done!!!! Add it to the list!
		orders[i] = currOrder
	}

	return
}
