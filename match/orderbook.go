package match

// LimitOrderbook provides all of the operations needed for a normal exchange orderbook. This should be a single-pair orderbook. TODO: Determine if this should actually be a single-pair orderbook
type LimitOrderbook interface {
	// SetMatchingAlgorithm should set the matching algorithm. TODO: I want this to be a thing.
	// SetMatchingAlgorithm(func(book *LimitOrderbook) (err error)) (err error)
	// PlaceOrders places multiple orders in the orderbook, and returns orders with their assigned IDs. These orders will all get the same time priority, we assume they come in at the same time.
	// TODO: figure out if this is the right thing to do, or if PlaceOrder should be here instead
	PlaceOrders(orders []*LimitOrder) (idOrders []*LimitOrder, err error)
	// GetBook takes in a trading pair and returns the whole orderbook.
	GetBook() (book []*LimitOrder, err error)
	// GetOrder gets an order from an OrderID
	GetOrder(id *OrderID) (order *LimitOrder, err error)
	// CancelOrder cancels an order with order id
	CancelOrder(id *OrderID) (err error)
	// CalculatePrice returns the calculated price based on the order book.
	CalculatePrice() (price Price, err error)
	// GetPairs gets the trading pair we can trade on
	GetPair() (pair *Pair, err error)
}
