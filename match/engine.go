package match

// The LimitEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
// One of these should be made for every pair.
type LimitEngine interface {
	PlaceLimitOrder(order *LimitOrder) (idRes *LimitOrderIDPair, err error)
	CancelLimitOrder(id *OrderID) (cancelled *CancelledOrder, err error)
	MatchLimitOrders() (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}

// The AuctionEngine is the interface for the internal matching engine. This should be the lowest level
// interface for the representation of a matching engine.
// One of these should be made for every pair.
type AuctionEngine interface {
	PlaceAuctionOrder(order *AuctionOrder, auctionID *AuctionID) (idRes *AuctionOrderIDPair, err error)
	CancelAuctionOrder(id *OrderID) (cancelled *CancelledOrder, cancelSettlement *SettlementExecution, err error)
	MatchAuctionOrders(auctionID *AuctionID) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error)
}

// SettlementEngine is an interface for something that keeps track of balances for users for a
// certain asset.
// One of these should be made for every asset.
type SettlementEngine interface {
	// ApplySettlementExecution is a method that applies a settlement execution.
	ApplySettlementExecution(setExec *SettlementExecution) (setRes *SettlementResult, err error)
	// CheckValid is a method that returns true if the settlement execution would be valid.
	CheckValid(setExec *SettlementExecution) (valid bool, err error)
}
