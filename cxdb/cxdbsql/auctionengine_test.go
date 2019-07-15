package cxdbsql

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
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

func TestCreateEngineAllParams(t *testing.T) {
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

	var ae *SQLAuctionEngine
	for _, pair := range pairList {
		if ae, err = CreateAucEngineStructWithConf(pair, testConfig()); err != nil {
			t.Errorf("Error creating auction engine for pair: %s", err)
		}
		if err = ae.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for auction engine: %s", err)
		}
	}

}

func TestPlaceSingleAuctionOrder(t *testing.T) {
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

	var ae *SQLAuctionEngine
	if ae, err = CreateAucEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		t.Errorf("Error creating auction engine for pair: %s", err)
	}

	defer func() {

		if err = ae.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for auction engine: %s", err)
		}

	}()

	var engine match.AuctionEngine = ae

	var idStruct *match.AuctionID = new(match.AuctionID)
	if err = idStruct.UnmarshalBinary(testEncryptedOrder.IntendedAuction[:]); err != nil {
		t.Errorf("Error unmarshalling auction ID: %s", err)
	}

	if _, err = engine.PlaceAuctionOrder(testAuctionOrder, idStruct); err != nil {
		t.Errorf("Error placing auction order: %s", err)
	}

}

func TestPlace1KAuctionOrders(t *testing.T) {
	PlaceMatchNAuctionOrdersTest(1000, t)
	return
}

func TestPlace10KAuctionOrders(t *testing.T) {
	PlaceMatchNAuctionOrdersTest(10000, t)
	return
}

func BenchmarkAllAuction(b *testing.B) {

	for _, howMany := range []uint64{1, 10, 100, 1000} {
		PlaceNAuctionOrders(howMany, b)
		MatchNAuctionOrders(howMany, b)
	}

}

func PlaceNAuctionOrders(howMany uint64, b *testing.B) {
	var err error

	var idStruct *match.AuctionID = new(match.AuctionID)
	if err = idStruct.UnmarshalBinary(testEncryptedOrder.IntendedAuction[:]); err != nil {
		b.Errorf("Error unmarshalling auction ID: %s", err)
	}

	var ordersToPlace []*match.AuctionOrder
	if ordersToPlace, err = fuzzManyOrders(howMany, testEncryptedOrder.IntendedPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		b.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.KillUser(); err != nil {
			b.Errorf("Error killing tester container user: %s", err)
		}
		if err = tc.CloseHandler(); err != nil {
			b.Errorf("Error killing tester container user: %s", err)
		}
	}()

	b.Run(fmt.Sprintf("Place%dAuctionOrders", howMany), func(b *testing.B) {

		defer func() {
			// drop here because we need to make sure the next run has a clean DB
			// the user persists because we designed the testerContainer to allow for that
			if err = tc.DropDBs(); err != nil {
				b.Errorf("Error killing tester container dbs: %s", err)
			}

		}()

		var ae *SQLAuctionEngine
		if ae, err = CreateAucEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
			b.Errorf("Error creating auction engine for pair: %s", err)
		}

		defer func() {

			if err = ae.DestroyHandler(); err != nil {
				b.Errorf("Error destroying handler for auction engine: %s", err)
			}

		}()

		var engine match.AuctionEngine = ae

		// Start it back up again, let's time this
		b.ResetTimer()

		for _, order := range ordersToPlace {
			if _, err = engine.PlaceAuctionOrder(order, idStruct); err != nil {
				b.Errorf("Error placing auction order: %s", err)
			}
		}

	})

}

func MatchNAuctionOrders(howMany uint64, b *testing.B) {
	var err error

	var idStruct *match.AuctionID = new(match.AuctionID)
	if err = idStruct.UnmarshalBinary(testEncryptedOrder.IntendedAuction[:]); err != nil {
		b.Errorf("Error unmarshalling auction ID: %s", err)
	}

	var ordersToPlace []*match.AuctionOrder
	if ordersToPlace, err = fuzzManyOrders(howMany, testEncryptedOrder.IntendedPair); err != nil {
		b.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		b.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.KillUser(); err != nil {
			b.Errorf("Error killing tester container user: %s", err)
		}
		if err = tc.CloseHandler(); err != nil {
			b.Errorf("Error killing tester container user: %s", err)
		}
	}()

	b.Run(fmt.Sprintf("Match%dAuctionOrders", howMany), func(b *testing.B) {

		defer func() {
			// drop here because we need to make sure the next run has a clean DB
			// the user persists because we designed the testerContainer to allow for that
			if err = tc.DropDBs(); err != nil {
				b.Errorf("Error killing tester container dbs: %s", err)
			}

		}()

		var ae *SQLAuctionEngine
		if ae, err = CreateAucEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
			b.Errorf("Error creating auction engine for pair: %s", err)
		}

		defer func() {

			if err = ae.DestroyHandler(); err != nil {
				b.Errorf("Error destroying handler for auction engine: %s", err)
			}

		}()

		var engine match.AuctionEngine = ae

		for _, order := range ordersToPlace {
			if _, err = engine.PlaceAuctionOrder(order, idStruct); err != nil {
				b.Errorf("Error placing auction order: %s", err)
			}
		}

		// Start it back up again, let's time this
		b.ResetTimer()

		if _, _, err = engine.MatchAuctionOrders(idStruct); err != nil {
			b.Errorf("Error matching auction orders: %s", err)
		}

		if err = tc.DropDBs(); err != nil {
			b.Errorf("Error killing tester container dbs: %s", err)
		}

	})

}

func PlaceMatchNAuctionOrdersTest(howMany uint64, t *testing.T) {
	var err error

	t.Logf("%s: Starting test setup", time.Now())
	var idStruct *match.AuctionID = new(match.AuctionID)
	if err = idStruct.UnmarshalBinary(testEncryptedOrder.IntendedAuction[:]); err != nil {
		t.Errorf("Error unmarshalling auction ID: %s", err)
	}

	var ordersToPlace []*match.AuctionOrder
	if ordersToPlace, err = fuzzManyOrders(howMany, testEncryptedOrder.IntendedPair); err != nil {
		t.Errorf("Error fuzzing many orders: %s", err)
	}

	var tc *testerContainer
	if tc, err = CreateTesterContainer(); err != nil {
		t.Errorf("Error creating tester container: %s", err)
	}

	defer func() {
		if err = tc.Kill(); err != nil {
			t.Errorf("Error killing tester container: %s", err)
		}
	}()

	var ae *SQLAuctionEngine
	if ae, err = CreateAucEngineStructWithConf(&testEncryptedOrder.IntendedPair, testConfig()); err != nil {
		t.Errorf("Error creating auction engine for pair: %s", err)
	}

	defer func() {

		if err = ae.DestroyHandler(); err != nil {
			t.Errorf("Error destroying handler for auction engine: %s", err)
		}

	}()

	var engine match.AuctionEngine = ae

	t.Logf("%s: Starting to place orders", time.Now())
	for _, order := range ordersToPlace {
		if _, err = engine.PlaceAuctionOrder(order, idStruct); err != nil {
			t.Errorf("Error placing auction order: %s", err)
		}
	}

	t.Logf("%s: Starting to match orders", time.Now())
	if _, _, err = engine.MatchAuctionOrders(idStruct); err != nil {
		t.Errorf("Error matching auction orders: %s", err)
	}
	t.Logf("%s: Done matching orders", time.Now())

	if err = tc.DropDBs(); err != nil {
		t.Errorf("Error killing tester container dbs: %s", err)
	}

}

// fuzzManyOrders creates a bunch of orders that have some random amounts, to make the clearing price super random
// It creates a new private key for every order, signs it after constructing it, etc.
func fuzzManyOrders(howMany uint64, pair match.Pair) (orders []*match.AuctionOrder, err error) {
	// want this to be seeded so it's reproducible

	// 21 million is max amount
	maxAmount := int64(2100000000000000)
	r := rand.New(rand.NewSource(1801))
	orders = make([]*match.AuctionOrder, howMany)
	for i := uint64(0); i < howMany; i++ {
		currOrder := new(match.AuctionOrder)
		var ephemeralPrivKey *koblitz.PrivateKey
		if ephemeralPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			err = fmt.Errorf("Error generating new private key: %s", err)
			return
		}
		currOrder.AmountWant = uint64(r.Int63n(maxAmount) + 1)
		currOrder.AmountHave = uint64(r.Int63n(maxAmount) + 1)
		if _, err = r.Read(currOrder.Nonce[:]); err != nil {
			err = fmt.Errorf("Cannot read random order nonce for FuzzManyOrders: %s", err)
			return
		}
		if r.Int63n(2) == 0 {
			currOrder.Side = "buy"
		} else {
			currOrder.Side = "sell"
		}
		currOrder.AuctionID = [32]byte{0xde, 0xad, 0xbe, 0xef}
		currOrder.TradingPair = match.Pair{
			AssetWant: btcreg,
			AssetHave: litereg,
		}
		pubkeyBytes := ephemeralPrivKey.PubKey().SerializeCompressed()
		copy(currOrder.Pubkey[:], pubkeyBytes)

		// now we hash the order
		orderBytes := currOrder.SerializeSignable()
		hasher := sha3.New256()
		hasher.Write(orderBytes)

		var sig *koblitz.Signature
		if sig, err = ephemeralPrivKey.Sign(hasher.Sum(nil)); err != nil {
			err = fmt.Errorf("Error signing order: %s", err)
			return
		}
		sigBytes := sig.Serialize()
		currOrder.Signature = make([]byte, len(sigBytes))
		copy(currOrder.Signature[:], sigBytes)
		// It's all done!!!! Add it to the list!
		orders[i] = currOrder
	}

	return
}
