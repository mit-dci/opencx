package match

import "fmt"

// Pair is a struct that represents a trading pair
type Pair struct {
	AssetWant byte `json:"assetWant"`
	// the AssetHave asset will always be the asset whose balance is checked
	AssetHave byte `json:"assetHave"`
}

// Order is a struct that represents a stored side of a trade
type Order interface {
	Type() string
	Price() (float64, error)
}

// LimitOrder represents a limit order, implementing the order interface
type LimitOrder struct {
	Client      string `json:"username"`
	Side        string `json:"side"`
	TradingPair Pair   `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amount"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64 `json:"price"`
}

// IsBuySide returns true if the limit order is buying
func (l *LimitOrder) IsBuySide() bool {
	return l.Side == "buy"
}

// IsSellSide returns true if the limit order is selling
func (l *LimitOrder) IsSellSide() bool {
	return l.Side == "sell"
}

// Price gets a float price for the order. This determines how it will get matched. The exchange should figure out if it can take some of the
// pennies off the dollar for things that request a certain amount but the amount they get (according to price and what the other side would be willing
// to give) is less than they officially requested. But tough luck to them we're taking fees anyways
func (l *LimitOrder) Price() (float64, error) {
	if l.AmountWant == 0 {
		return 0, fmt.Errorf("The amount requested in the order is 0, so no price can be calculated. Consider it a donation")
	}
	return float64(l.AmountHave / l.AmountWant), nil
}

// generateUniquePairs generates unique asset pairs based on the assets available
func generateUniquePairs(assetList []byte) []Pair {

	assetListLen := len(assetList)
	numPairIndeces := assetListLen * (assetListLen - 1) / 2
	var pairList = make([]Pair, numPairIndeces)
	pairListIndex := 0
	for i, elem := range assetList {
		for lower := i + 1; lower < assetListLen; lower++ {
			pairList[pairListIndex].AssetWant = elem
			pairList[pairListIndex].AssetHave = assetList[lower]
			pairListIndex++
		}
	}

	return pairList
}

// GenerateAssetPairs generates unique asset pairs based on the default assets available
func GenerateAssetPairs() []Pair {
	return generateUniquePairs(AssetList())
}

// String is the tostring function for a pair
func (p Pair) String() string {
	return ByteToAssetString(p.AssetWant) + "/" + ByteToAssetString(p.AssetHave)
}
