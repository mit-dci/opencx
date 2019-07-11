package match

import "time"

// AuctionBatch is a struct that represents a batch of auction order results
type AuctionBatch struct {
	Batch     []*OrderPuzzleResult
	AuctionID [32]byte
}

// AuctionBatcher is an interface for a service that collects orders and handles batching per auction.
// This is abstracted because solving puzzles is a task that should be easily outsourceable, and should
// not be integrated into the core logic. One could easily see a server that performs puzzle solving
// that is separate from the actual exchange. The exchange doesn't need to schedule puzzle solving,
// or worry about scaling it, but the auction batcher does. The auction batcher needs to involve solving
// many puzzles at once.
type AuctionBatcher interface {
	// RegisterAuction registers a new auction with a specified Auction ID, which will be an array of
	// 32 bytes.
	RegisterAuction(auctionID [32]byte) (err error)

	// AddEncrypted adds an encrypted order to an auction. This should error if either the auction doesn't
	// exist, or the auction is ended.
	AddEncrypted(order *EncryptedAuctionOrder) (err error)

	// EndAuction ends the auction with the specified auction ID, and returns the channel which will
	// receive a batch of orders puzzle results. This is like a promise. This channel should be of size 1.
	EndAuction(auctionID [32]byte) (batchChan chan *AuctionBatch, err error)

	// ActiveAuctions returns a map of auction id to time TODO: figure out if this is really necessary
	ActiveAuctions() (activeBatches map[[32]byte]time.Time)
}

// BatchResult is a struct that represents the result of a batch auction.
type BatchResult struct {
	OriginalBatch *AuctionBatch
	// RejectedResults and AcceptedResults should be disjoint sets
	RejectedResults []*OrderPuzzleResult
	AcceptedResults []*OrderPuzzleResult
}
