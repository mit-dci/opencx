package match

// CancelledOrder is broadcasted from the matching engine, marking an order as cancelled.
type CancelledOrder struct {
	OrderID *OrderID
	// Debited is part of CancelledOrder because when an order is cancelled, assets are freed for use in
	// other orders.
	Debited *Entry
}
