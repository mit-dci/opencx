package cxdb

import (
	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

// SettlementStore is like a frontend for settlements, it contains balances that are readable
type SettlementStore interface {
	// UpdateBalances updates the balances from the settlement executions
	UpdateBalances(settlementExecs []*match.SettlementResult) (err error)
	// GetBalance gets the balance for a pubkey and an asset.
	GetBalance(pubkey *koblitz.PublicKey) (balance uint64, err error)
}

type DepositStore interface {
	// RegisterUser takes in a pubkey, and an address for the pubkey
	RegisterUser(pubkey *koblitz.PublicKey, address string) (err error)
	// UpdateDeposits updates the deposits when a block comes in
	UpdateDeposits(deposits []match.Deposit, blockheight uint64) (depositExecs []*match.SettlementExecution, err error)
	// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
	GetDepositAddressMap() (depAddrMap map[string]*koblitz.PublicKey, err error)
	// GetDepositAddress gets the deposit address for a pubkey and an asset.
	GetDepositAddress(pubkey *koblitz.PublicKey) (addr string, err error)
}

// PuzzleStore is an interface for defining a storage layer for auction order puzzles.
type PuzzleStore interface {
	// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
	// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
	// what was submitted.
	ViewAuctionPuzzleBook(auctionID *match.AuctionID) (puzzles []*match.EncryptedAuctionOrder, err error)
	// PlaceAuctionPuzzle puts an encrypted auction order in the datastore.
	PlaceAuctionPuzzle(puzzledOrder *match.EncryptedAuctionOrder) (err error)
}
