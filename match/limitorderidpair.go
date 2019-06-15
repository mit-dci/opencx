package match

import "time"

// LimitOrderIDPair is order ID, order, and time, used for generating executions in limit order matching algorithms
type LimitOrderIDPair struct {
	Timestamp time.Time
	OrderID   *OrderID
	Order     *LimitOrder
}
