package cxauctionserver

import (
	"fmt"
	"sync"

	"github.com/mit-dci/opencx/crypto/timelockencoders"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// intermediateBatch is used to manage status of batch auctions.
type intermediateBatch struct {
	orderChan    chan *match.OrderPuzzleResult
	solvedOrders []*match.OrderPuzzleResult
	solvedChan   chan *match.AuctionBatch
	id           [32]byte
	// numOrders is the number of orders left to be solved
	numOrders      uint64
	active         bool
	orderUpdateMtx sync.Mutex
	offChan        chan bool
}

// ABather is a very simple, small scale, non-persistent batcher.
type ABatcher struct {
	batchMap    map[[32]byte]*intermediateBatch
	batchMapMtx sync.Mutex
}

// NewABatcher creates a new AuctionBatcher.
func NewABatcher() (batcher *ABatcher, err error) {
	batcher = &ABatcher{
		batchMap:    make(map[[32]byte]*intermediateBatch),
		batchMapMtx: sync.Mutex{},
	}
	return
}

// RegisterAuction registers a new auction with a specified Auction ID, which will be an array of
// 32 bytes.
func (ab *ABatcher) RegisterAuction(auctionID [32]byte) (err error) {
	if ab.batchMap == nil {
		err = fmt.Errorf("Cannot send order to batcher that isn't set up")
		return
	}
	ab.batchMapMtx.Lock()
	var thisBatch *intermediateBatch
	thisBatch = &intermediateBatch{
		orderChan:      make(chan *match.OrderPuzzleResult),
		solvedOrders:   []*match.OrderPuzzleResult{},
		solvedChan:     make(chan *match.AuctionBatch, 1),
		id:             auctionID,
		numOrders:      0,
		active:         true,
		orderUpdateMtx: sync.Mutex{},
		offChan:        make(chan bool, 1),
	}
	ab.batchMap[auctionID] = thisBatch
	ab.batchMapMtx.Unlock()
	return
}

// orderSolver spawns an order solver. This should be done in a goroutine.
func (ib *intermediateBatch) orderSolver() {
	var currResult *match.OrderPuzzleResult
	for {
		select {
		case <-ib.offChan:
			close(ib.offChan)
			close(ib.orderChan)
			return
		case currResult = <-ib.orderChan:
			// Lock things. minus numOrders.
			// Check if it's 0 and active is false.
			// If so, add to the solved channel.
			ib.orderUpdateMtx.Lock()
			// invariant: if numOrders is already 0 and we try to minus it. I think we've covered this
			// case with everything already but just noting it down here.
			ib.numOrders--
			ib.solvedOrders = append(ib.solvedOrders, currResult)
			if !ib.active && ib.numOrders == 0 {
				// put the batch into the channel
				ib.solvedChan <- &match.AuctionBatch{
					Batch:     ib.solvedOrders,
					AuctionID: ib.id,
				}
				// turn this off, clean up
				ib.offChan <- true
			}
			ib.orderUpdateMtx.Unlock()
		default:
		}
	}
	return
}

// AddEncrypted adds an encrypted order to an auction. This should error if either the auction doesn't
// exist, or the auction is ended.
func (ab *ABatcher) AddEncrypted(order *match.EncryptedAuctionOrder) (err error) {
	// First retreive the interBatch, this shouldn't change but we use a mutex to access the map.
	ab.batchMapMtx.Lock()

	// Check that we can do things to the batch
	var interBatch *intermediateBatch
	var ok bool
	if interBatch, ok = ab.batchMap[order.IntendedAuction]; !ok {
		err = fmt.Errorf("Cannot add encrypted order to unregistered auction")
		ab.batchMapMtx.Unlock()
		return
	}

	ab.batchMapMtx.Unlock()

	// now we have this other mutex because we are going to be checking and modifying the contents
	interBatch.orderUpdateMtx.Lock()
	if !interBatch.active {
		err = fmt.Errorf("Cannot add encrypted order to inactive auction")
		interBatch.orderUpdateMtx.Unlock()
		return
	}

	interBatch.numOrders++

	go solveSingleOrder(order, interBatch.orderChan)

	interBatch.orderUpdateMtx.Unlock()
	return
}

// solveSingleOrder solves a single order and deposits the result into the sendResChan. This should
// be done in a goroutine.
func solveSingleOrder(eOrder *match.EncryptedAuctionOrder, sendResChan chan *match.OrderPuzzleResult) {
	var err error
	result := new(match.OrderPuzzleResult)
	result.Encrypted = eOrder

	// send to channel at end of method
	defer func() {
		// Make sure we can actually send to this channel
		logging.Infof("sendResChan cap: %d, len: %d", cap(sendResChan), len(sendResChan))
		select {
		case sendResChan <- result:
			return
		default:
			panic("Couldn't send result to channel! panicking!")
		}
	}()

	var orderBytes []byte
	if orderBytes, err = timelockencoders.SolvePuzzleRC5(eOrder.OrderCiphertext, eOrder.OrderPuzzle); err != nil {
		result.Err = fmt.Errorf("Error solving RC5 puzzle for solve single order: %s", err)
		return
	}

	result.Auction = new(match.AuctionOrder)
	if err = result.Auction.Deserialize(orderBytes); err != nil {
		result.Err = fmt.Errorf("Error deserializing order from puzzle for solve single order: %s", err)
		return
	}

	return
}

// EndAuction ends the auction with the specified auction ID, and returns the channel which will
// receive a batch of orders puzzle results. This is like a promise. This channel should be of size 1.
// TODO: add commitment to this?
func (ab *ABatcher) EndAuction(auctionID [32]byte) (batchChan chan *match.AuctionBatch, err error) {
	// First retreive the interBatch, this shouldn't change but we use a mutex to access the map.
	ab.batchMapMtx.Lock()

	// Check that we can do things to the batch
	var interBatch *intermediateBatch
	var ok bool
	if interBatch, ok = ab.batchMap[auctionID]; !ok {
		err = fmt.Errorf("Cannot end unregistered auction")
		ab.batchMapMtx.Unlock()
		return
	}

	ab.batchMapMtx.Unlock()

	// now we have this other mutex because we are going to be checking and modifying the contents
	interBatch.orderUpdateMtx.Lock()
	if !interBatch.active {
		err = fmt.Errorf("Cannot end inactive auction")
		interBatch.orderUpdateMtx.Unlock()
		return
	}
	interBatch.active = false
	batchChan = interBatch.solvedChan
	interBatch.orderUpdateMtx.Unlock()
	return
}
