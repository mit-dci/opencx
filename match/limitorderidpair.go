package match

import "time"

// LimitOrderIDPair is order ID, order, price, and time, used for generating executions in limit order matching algorithms
type LimitOrderIDPair struct {
	Timestamp time.Time   `json:"timestamp"`
	Price     float64     `json:"price"`
	OrderID   *OrderID    `json:"orderid"`
	Order     *LimitOrder `json:"limitorder"`
}
