package cxdb

import (
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

// OpencxStore should define all the functions that are required for an implementation of an exchange that uses any sort of I/O to store its information.
// If you wanted you could implement everything entirely in golang's allocated memory, with built in data structures (not that it is advised). The datastore layer
// should not check validity of the things it is doing, just update or insert or return whatever.
type OpencxStore interface {
	// SetupClient makes sure that whatever things need to be done before we use the datastore can be done before we need to use the datastore.
	SetupClient([]*coinparam.Params) error
	// RegisterUser takes in a pubkey, and a map of asset to addresses for the pubkey. It inserts the necessary information in databases to register the pubkey.
	RegisterUser(*koblitz.PublicKey, map[*coinparam.Params]string) error
	// GetBalance gets the balance for a pubkey and an asset.
	GetBalance(*koblitz.PublicKey, *coinparam.Params) (uint64, error)
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
	Withdraw(*koblitz.PublicKey, *coinparam.Params, uint64) error
	// UpdateDeposits updates the deposits when a block comes in
	UpdateDeposits([]match.Deposit, uint64, *coinparam.Params) error
	// AddToBalance adds to the balance of a user
	AddToBalance(*koblitz.PublicKey, uint64, *coinparam.Params) error
	// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
	GetDepositAddressMap(*coinparam.Params) (map[string]*koblitz.PublicKey, error)
	// GetOrder gets an order from an OrderID
	GetOrder(string) (*match.LimitOrder, error)
	// GetOrdersForPubkey gets orders for a specific pubkey
	GetOrdersForPubkey(*koblitz.PublicKey) ([]*match.LimitOrder, error)
	// CancelOrder cancels an order with order id
	CancelOrder(string) error
}

// TODO: separate out parts of the Store, like many of the account based operations (balance and whatnot), so
// an auction server and exchange server could share the same account store, but different order stores

// OpencxAuctionStore should define all the functions that are required for an implementation of a
// front-running resistant auction exchange that uses any sort of I/O to store its information.
// If you wanted you could implement everything entirely in golang's allocated memory, with built in data structures (not that it is advised). The datastore layer
// should not check validity of the things it is doing, just update or insert or return whatever.
type OpencxAuctionStore interface {
	// SetupClient makes sure that whatever things need to be done before we use the datastore can be done before we need to use the datastore.
	SetupClient([]*coinparam.Params) error
	// RegisterUser takes in a pubkey, and a map of asset to addresses for the pubkey. It inserts the necessary information in databases to register the pubkey.
	RegisterUser(*koblitz.PublicKey, map[*coinparam.Params]string) error
	// GetBalance gets the balance for a pubkey and an asset.
	GetBalance(*koblitz.PublicKey, *coinparam.Params) (uint64, error)
	// Withdraw checks the user's balance against the amount and if valid, reduces the balance by that amount.
	Withdraw(*koblitz.PublicKey, *coinparam.Params, uint64) error
	// AddToBalance adds to the balance of a user
	AddToBalance(*koblitz.PublicKey, uint64, *coinparam.Params) error
	// PlaceAuctionPuzzle puts an encrypted auction order in the datastore.
	PlaceAuctionPuzzle(*match.EncryptedAuctionOrder) error
	// PlaceAuctionOrder places an order in the unencrypted datastore.
	PlaceAuctionOrder(*match.AuctionOrder) error
	// ViewAuctionOrderBook takes in a trading pair and auction ID, and returns auction orders.
	ViewAuctionOrderBook(*match.Pair, [32]byte) ([]*match.AuctionOrder, []*match.AuctionOrder, error)
	// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
	// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
	// what was submitted.
	ViewAuctionPuzzleBook([32]byte) ([]*match.EncryptedAuctionOrder, error)
}
