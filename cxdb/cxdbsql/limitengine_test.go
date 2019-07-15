package cxdbsql

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

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

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		t.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			t.Errorf("Error killing tester container: %s", err)
		}
	}()

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(constCoinParams()); err != nil {
		t.Errorf("Error creating asset pairs from coin list: %s", err)
	}

	var le *SQLLimitEngine
	for _, pair := range pairList {
		if le, err = CreateLimEngineStructWithConf(pair, testConfig()); err != nil {
			t.Errorf("Error creating limit engine for pair: %s", err)
		}

		if err = le.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}

}

func TestPlaceSingleLimitOrder(t *testing.T) {
	var err error

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		t.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			t.Errorf("Error killing tester container: %s", err)
		}
	}()

	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		t.Errorf("Error creating limit engine for pair: %s", err)
	}

	defer func() {
		if err = le.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}()
	var engine match.LimitEngine = le

	if _, err = engine.PlaceLimitOrder(testLimitOrder); err != nil {
		t.Errorf("Error placing limit order: %s", err)
	}

}

func TestPlaceMatch1KLimitOrders(t *testing.T) {
	PlaceMatchNLimitOrdersTest(1000, t)
	return
}

func TestPlaceMatch2KLimitOrders(t *testing.T) {
	PlaceMatchNLimitOrdersTest(2000, t)
	return
}

func BenchmarkAllLimit(b *testing.B) {

	for _, howMany := range []uint64{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("Place%dLimitOrders", howMany), func(b *testing.B) {
			PlaceNLimitOrders(howMany, b)
		})
		b.Run(fmt.Sprintf("Match%dLimitOrders", howMany), func(b *testing.B) {
			MatchNLimitOrders(howMany, b)
		})
		b.Run(fmt.Sprintf("PlaceMatch%dLimitOrders", howMany), func(b *testing.B) {
			PlaceMatchNLimitOrders(howMany, b)
		})
	}

}

func BenchmarkPlaceMatch1LimitOrders(b *testing.B) {
	PlaceMatchNLimitOrders(1, b)
	return
}

func BenchmarkPlaceMatch10LimitOrders(b *testing.B) {
	PlaceMatchNLimitOrders(10, b)
	return
}

func BenchmarkPlaceMatch100LimitOrders(b *testing.B) {
	PlaceMatchNLimitOrders(100, b)
	return
}

func BenchmarkPlaceMatch1000LimitOrders(b *testing.B) {
	PlaceMatchNLimitOrders(1000, b)
	return
}

func BenchmarkMatch1LimitOrders(b *testing.B) {
	MatchNLimitOrders(1, b)
	return
}

func BenchmarkMatch10LimitOrders(b *testing.B) {
	MatchNLimitOrders(10, b)
	return
}

func BenchmarkMatch100LimitOrders(b *testing.B) {
	MatchNLimitOrders(100, b)
	return
}

func BenchmarkMatch1000LimitOrders(b *testing.B) {
	MatchNLimitOrders(1000, b)
	return
}

func BenchmarkPlace1LimitOrders(b *testing.B) {
	PlaceNLimitOrders(1, b)
	return
}

func BenchmarkPlace10LimitOrders(b *testing.B) {
	PlaceNLimitOrders(10, b)
	return
}

func BenchmarkPlace100LimitOrders(b *testing.B) {
	PlaceNLimitOrders(100, b)
	return
}

func BenchmarkPlace1000LimitOrders(b *testing.B) {
	PlaceNLimitOrders(1000, b)
	return
}

func PlaceNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		b.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			b.Errorf("Error killing tester container: %s", err)
		}
	}()

	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		b.Errorf("Error creating limit engine for pair: %s", err)
	}

	defer func() {
		if err = le.DestroyHandler(); err != nil {
			b.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}()

	var engine match.LimitEngine = le

	// Start it back up again, let's time this
	b.ResetTimer()

	for _, order := range ordersToPlace {
		if order == nil {
			b.Errorf("Order is nil?? this should not happen")
		}
		if engine == nil {
			b.Errorf("Engine is nil?? this should not happen")
		}
		if _, err = engine.PlaceLimitOrder(order); err != nil {
			b.Errorf("Error placing limit order: %s", err)
		}
	}

}

func MatchNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		b.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			b.Errorf("Error killing tester container: %s", err)
		}
	}()

	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		b.Errorf("Error creating limit engine for pair: %s", err)
	}

	defer func() {
		if err = le.DestroyHandler(); err != nil {
			b.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}()
	var engine match.LimitEngine = le

	for _, order := range ordersToPlace {
		if _, err = engine.PlaceLimitOrder(order); err != nil {
			b.Errorf("Error placing limit order: %s", err)
		}
	}

	// Start it back up again, let's time this
	b.ResetTimer()

	if _, _, err = engine.MatchLimitOrders(); err != nil {
		b.Errorf("Error matching limit orders: %s", err)
	}

}

// PlaceMatchNLimitOrders is a bit different but more real, it places an order then runs matching. This should get an idea of overall throughput, maybe cause some errors
func PlaceMatchNLimitOrders(howMany uint64, b *testing.B) {
	var err error

	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		b.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			b.Errorf("Error killing tester container: %s", err)
		}
	}()

	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		b.Errorf("Error creating limit engine for pair: %s", err)
	}

	defer func() {
		if err = le.DestroyHandler(); err != nil {
			b.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}()
	var engine match.LimitEngine = le

	// Start it back up again, let's time this
	b.ResetTimer()
	for _, order := range ordersToPlace {
		if _, err = engine.PlaceLimitOrder(order); err != nil {
			b.Errorf("Error placing limit order: %s", err)
		}
		if _, _, err = engine.MatchLimitOrders(); err != nil {
			b.Errorf("Error matching limit orders: %s", err)
		}
	}

}

// PlaceMatchNLimitOrdersTest is a bit different but more real, it places an order then runs matching. This should get an idea of overall throughput, maybe cause some errors
func PlaceMatchNLimitOrdersTest(howMany uint64, t *testing.T) {
	var err error

	t.Logf("%s: Starting test setup", time.Now())
	var ordersToPlace []*match.LimitOrder
	if ordersToPlace, err = fuzzManyLimitOrders(howMany, testLimitOrder.TradingPair); err != nil {
		t.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		t.Errorf("Error creating tester container: %s", err)
	}

	// Clean out database
	if err = tc.DropDBs(); err != nil {
		t.Errorf("Error making db clean for tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			t.Errorf("Error killing tester container: %s", err)
		}
	}()

	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		t.Errorf("Error creating limit engine for pair: %s", err)
	}

	defer func() {
		if err = le.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for limit engine: %s", err)
		}
	}()
	var engine match.LimitEngine = le

	t.Logf("%s: Starting to place and match orders", time.Now())
	for _, order := range ordersToPlace {
		if _, err = engine.PlaceLimitOrder(order); err != nil {
			t.Errorf("Error placing limit order: %s", err)
		}
		if _, _, err = engine.MatchLimitOrders(); err != nil {
			t.Errorf("Error matching limit orders: %s", err)
		}
	}
	t.Logf("%s: Done placing and matching orders", time.Now())

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
