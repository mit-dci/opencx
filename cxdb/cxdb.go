package cxdb

import (
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

type SettlementStore interface {
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
}

type DepositStore interface {
	// SetupClient makes sure that whatever things need to be done before we use the datastore can be done before we need to use the datastore.
	SetupClient([]*coinparam.Params) error
	// UpdateDeposits updates the deposits when a block comes in
	UpdateDeposits([]match.Deposit, uint64, *coinparam.Params) error
	// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
	GetDepositAddressMap(*coinparam.Params) (map[string]*koblitz.PublicKey, error)
	// GetDepositAddress gets the deposit address for a pubkey and an asset.
	GetDepositAddress(*koblitz.PublicKey, string) (string, error)
}

// PuzzleStore is an interface for defining a storage layer for auction order puzzles.
type PuzzleStore interface {
	// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
	// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
	// what was submitted.
	ViewAuctionPuzzleBook([32]byte) ([]*match.EncryptedAuctionOrder, error)
	// PlaceAuctionPuzzle puts an encrypted auction order in the datastore.
	PlaceAuctionPuzzle(*match.EncryptedAuctionOrder) error
}
