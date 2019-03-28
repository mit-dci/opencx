package cxdb

import (
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
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
	SetupClient([]*coinparam.Params) error
	// RegisterUser takes in a pubkey, and a map of asset to addresses for the pubkey. It inserts the necessary information in databases to register the pubkey.
	RegisterUser(*koblitz.PublicKey, map[*coinparam.Params]string) error
	// GetBalance gets the balance for a pubkey and an asset.
	GetBalance(*koblitz.PublicKey, string) (uint64, error)
	// GetDepositAddress gets the deposit address for a pubkey and an asset.
	GetDepositAddress(*koblitz.PublicKey, string) (string, error)
	// GetPairs gets all the trading pairs that we can trade on
	GetPairs() []*match.Pair
	// PlaceOrder places an order in the datastore.
	PlaceOrder(*match.LimitOrder) (string, error)
	// ViewOrderBook takes in a trading pair and returns buy order and sell orders separately. This should eventually only return the orderbook, not sure.
	ViewOrderBook(*match.Pair) ([]*match.LimitOrder, []*match.LimitOrder, error)
	// CalculatePrice returns the calculated price based on the order book.
	CalculatePrice(*match.Pair) (float64, error)
	// Withdraw checks the user's balance against the amount and if valid, reduces the balance by that amount.
	Withdraw(*koblitz.PublicKey, string, uint64) error
	// RunMatching runs the matching logic for all prices in the exchange. Matching is less abstracted because it is done very often, so the overhead makes a difference.
	RunMatching(*match.Pair) error
	// RunMatchingForPrice runs matching only for a specific price, likely the price that an order is coming in
	RunMatchingForPrice(*match.Pair, float64) error
	// UpdateDeposits updates the deposits when a block comes in
	UpdateDeposits([]match.Deposit, uint64, *coinparam.Params) error
	// LightningDeposit updates the balance of a user who is funding through lightning
	LightningDeposit(*koblitz.PublicKey, uint64, *coinparam.Params, uint32) error
	// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
	GetDepositAddressMap(*coinparam.Params) (map[string]*koblitz.PublicKey, error)
	// GetOrder gets an order from an OrderID
	GetOrder(string) (*match.LimitOrder, error)
	// GetOrdersForPubkey gets orders for a specific pubkey
	GetOrdersForPubkey(*koblitz.PublicKey) ([]*match.LimitOrder, error)
	// CancelOrder cancels an order with order id
	CancelOrder(string) error
}
