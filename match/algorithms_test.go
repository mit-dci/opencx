package match

import (
	"testing"

	"github.com/mit-dci/lit/coinparam"
)

var (
	litereg, _ = AssetFromCoinParam(&coinparam.LiteRegNetParams)
	btcreg, _  = AssetFromCoinParam(&coinparam.RegressionNetParams)
	BTC_LTC    = &Pair{
		AssetWant: btcreg,
		AssetHave: litereg,
	}
	onePriceSell = &AuctionOrder{
		Pubkey:      [33]byte{},
		Side:        "sell",
		TradingPair: *BTC_LTC,
		AmountWant:  1000,
		AmountHave:  1000,
	}
)

func TestClearingPriceSamePriceBook(t *testing.T) {

	return
}
