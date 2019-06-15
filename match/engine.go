package match

// The LimitEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
type LimitEngine interface {
	PlaceOrder(order *LimitOrder) (idRes *LimitOrderIDPair, err error)
	CancelOrder(id *OrderID) (cancelled *CancelledOrder, err error)
	MatchOrders() (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}

// The AuctionEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
type AuctionEngine interface {
	PlaceOrder(order *LimitOrder, auctionID *AuctionID) (idRes *AuctionOrderIDPair, err error)
	CancelOrder(id *OrderID) (cancelled *CancelledOrder, err error)
	MatchOrders(auctionID *AuctionID) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}
