package match

// Pair is a struct that represents a trading pair
type Pair struct {
	AssetWant byte `json:"assetWant"`
	// the AssetHave asset will always be the asset whose balance is checked
	AssetHave byte `json:"assetHave"`
}

// Order is a struct that represents a stored side of a trade
type Order struct {
	Client      string  `json:"username"`
	Side        string  `json:"side"`
	TradingPair Pair    `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave  int64   `json:"amount"`
	// amount of assetWant the user wants for their assetHave
	AmountWant  int64    `json:"price"`
}

func(o *Order) isBuySide() bool {
	return o.Side == "buy"
}

func(o *Order) isSellSide() bool {
	return o.Side == "sell"
}
