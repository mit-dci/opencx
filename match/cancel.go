package match

// CancelledOrder is broadcasted from the matching engine, marking an order as cancelled.
type CancelledOrder struct {
	OrderID *OrderID
}
