package match

import "time"

// LimitOrderIDPair is order ID, order, price, and time, used for generating executions in limit order matching algorithms
type LimitOrderIDPair struct {
	Timestamp time.Time
	Price     float64
	OrderID   *OrderID
	Order     *LimitOrder
}
