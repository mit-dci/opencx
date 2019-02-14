package cxdb

import (
	"github.com/mit-dci/opencx/match"
)

// abstractions to make:
// - user abstraction
// - asset abstraction
// - use real addresses
// - change limit order to generic order?? Different behavior based on type of order?
// like what if in matching it checked if it were a market order and just checked the amount put in, and satisfied the amount out on matching. That would work.

// OpencxStore should define all the functions that are required for an implementation of an exchange that uses any sort of I/O to store its information.
// If you wanted you could implement everything entirely in golang's allocated memory, with built in data structures (not that it is advised). The datastore layer
// should not check validity of the things it is doing, just update or insert or return whatever.
type OpencxStore interface {
	// SetupClient makes sure that whatever things need to be done before we use the datastore can be done before we need to use the datastore.
	SetupClient() error
	// RegisterUser takes in a user, and a map of asset to addresses for the user. It inserts the necessary information in databases to register the user.
	RegisterUser(string, map[string]string) error
	// GetBalance gets the balance for a user and an asset.
	GetBalance(string, string) (float64, error)
	// GetDepositAddress gets the deposit address for a user and an asset.
	GetDepositAddress(string, string) (string, error)
	// PlaceOrder places an order in the datastore.
	PlaceOrder(match.LimitOrder) error
	// ViewOrderBook takes in a trading pair and returns buy order and sell orders separately. This should eventually only return the orderbook, not sure.
	ViewOrderBook(match.Pair) ([]*match.LimitOrder, []*match.LimitOrder, error)
	// CalculatePrice returns the calculated price based on the order book.
	CalculatePrice(match.Pair) (float64, error)
	// Withdraw checks the user's balance against the amount and if valid, reduces the balance by that amount.
	Withdraw(string, string, uint64) error
}
