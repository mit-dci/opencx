package cxauctionserver

import "github.com/mit-dci/opencx/match"

type ABatcher struct {
	orderChanMap map[[32]byte]chan *match.AuctionBatch
}

// NewABatcher creates a new AuctionBatcher.
func NewABatcher() (batcher *ABatcher, err error) {
	batcher = &ABatcher{
		orderChanMap: make(map[[32]byte]chan *match.AuctionBatch),
	}
	return
}

// RegisterAuction registers a new auction with a specified Auction ID, which will be an array of
// 32 bytes.
func (ab *ABatcher) RegisterAuction(auctionID [32]byte) (err error) {
	// TODO
	return
}

// AddEncrypted adds an encrypted order to an auction. This should error if either the auction doesn't
// exist, or the auction is ended.
func (ab *ABatcher) AddEncrypted(order *match.EncryptedAuctionOrder) (err error) {
	// TODO
	return
}

// EndAuction ends the auction with the specified auction ID, and returns the channel which will
// receive a batch of orders puzzle results. This is like a promise. This channel should be of size 1.
func (ab *ABatcher) EndAuction(auctionID [32]byte) (batchChan chan *match.AuctionBatch, err error) {
	// TODO
	return
}
