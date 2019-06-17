package match

// The LimitEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
// One of these should be made for every pair.
type LimitEngine interface {
	PlaceOrder(order *LimitOrder) (idRes *LimitOrderIDPair, err error)
	CancelOrder(id *OrderID) (cancelled *CancelledOrder, err error)
	MatchOrders() (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}

// The AuctionEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
// One of these should be made for every pair.
type AuctionEngine interface {
	PlaceOrder(order *AuctionOrder, auctionID *AuctionID) (idRes *AuctionOrderIDPair, err error)
	CancelOrder(id *OrderID) (cancelled *CancelledOrder, err error)
	MatchOrders(auctionID *AuctionID) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}

// SettlementEngine is an interface for something that keeps track of balances for users for a
// certain asset.
// One of these should be made for every asset.
type SettlementEngine interface {
	// ApplySettlementExecution is a method that applies a settlement execution.
	ApplySettlementExecution(setExec *SettlementExecution) (err error)
	// CheckValid is a method that returns true if the settlement execution would be valid.
	CheckValid(setExec *SettlementExecution) (valid bool, err error)
}
