package match

// LimitBookStruct is a struct that represents a Limit Orderbook.
// It tries to be very quick to access and manipulate, so the
// orderbook is represented as an array.
type LimitBookStruct struct {
	PriceLevels []LimitQueue
}

// LimitQueue represents a time-ordered queue for buy and sell orders
type LimitQueue struct {
	BuyOrders  []LimitOrder
	SellOrders []LimitOrder
}
