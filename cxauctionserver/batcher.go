package cxauctionserver

import (
	"fmt"
	"sync"
	"time"

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
	maxOrders      uint64
	// just for display
	started time.Time
}

// ABather is a very simple, small scale, non-persistent batcher.
type ABatcher struct {
	batchMap     map[[32]byte]*intermediateBatch
	batchMapMtx  sync.Mutex
	maxBatchSize uint64
}

// NewABatcher creates a new AuctionBatcher.
func NewABatcher(maxBatchSize uint64) (batcher *ABatcher, err error) {
	batcher = &ABatcher{
		batchMap:     make(map[[32]byte]*intermediateBatch),
		batchMapMtx:  sync.Mutex{},
		maxBatchSize: maxBatchSize,
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
	if ab.maxBatchSize == 0 {
		err = fmt.Errorf("Cannot have a max batch size of 0")
		return
	}

	ab.batchMapMtx.Lock()
	var thisBatch *intermediateBatch
	thisBatch = &intermediateBatch{
		orderChan:      make(chan *match.OrderPuzzleResult, ab.maxBatchSize),
		solvedOrders:   []*match.OrderPuzzleResult{},
		solvedChan:     make(chan *match.AuctionBatch, 1),
		id:             auctionID,
		numOrders:      0,
		active:         true,
		orderUpdateMtx: sync.Mutex{},
		offChan:        make(chan bool, 1),
		maxOrders:      ab.maxBatchSize,
		started:        time.Now(),
	}
	ab.batchMap[auctionID] = thisBatch

	// Start the solver
	go thisBatch.OrderSolver()

	ab.batchMapMtx.Unlock()
	return
}

// ActiveAuctions returns a map of auction id to time
func (ab *ABatcher) ActiveAuctions() (activeBatches map[[32]byte]time.Time) {
	activeBatches = make(map[[32]byte]time.Time)

	ab.batchMapMtx.Lock()

	for id, batch := range ab.batchMap {
		if batch.active {
			activeBatches[id] = batch.started
		}
	}

	ab.batchMapMtx.Unlock()

	return
}

// orderSolver spawns an order solver. This should be done in a goroutine.
func (ib *intermediateBatch) OrderSolver() {
	var currResult *match.OrderPuzzleResult
	for i := ib.maxOrders; i > 0; i-- {
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
		}
	}
	return
}

// solveSingleOrder solves a single order and deposits the result into the sendResChan. This should
// be done in a goroutine.
func (ib *intermediateBatch) solveSingleOrder(eOrder *match.EncryptedAuctionOrder) {
	var err error
	result := new(match.OrderPuzzleResult)
	result.Encrypted = eOrder

	logging.Infof("Start of solve, num orders left: %d", ib.numOrders)

	// send to channel at end of method
	defer func() {
		// Make sure we can actually send to this channel
		logging.Infof("Try send, num orders left: %d", ib.numOrders)
		select {
		case ib.orderChan <- result:
			logging.Infof("Sent order to channel")
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

	go interBatch.solveSingleOrder(order)

	interBatch.orderUpdateMtx.Unlock()
	return
}

// EndAuction ends the auction with the specified auction ID, and returns the channel which will
// receive a batch of orders puzzle results. This is like a promise. This channel should be of size 1.
// TODO: add commitment to this?
func (ab *ABatcher) EndAuction(auctionID [32]byte) (batchChan chan *match.AuctionBatch, err error) {
	ab.batchMapMtx.Lock()

	// First retreive the interBatch, this shouldn't change but we use a mutex to access the map.
	// Check that we can do things to the batch
	var interBatch *intermediateBatch
	var ok bool
	if interBatch, ok = ab.batchMap[auctionID]; !ok {
		err = fmt.Errorf("Cannot end unregistered auction %x", auctionID)
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

func CreateAuctionBatcherMap(pairList []*match.Pair, maxBatchSize uint64) (batchers map[match.Pair]match.AuctionBatcher, err error) {
	batchers = make(map[match.Pair]match.AuctionBatcher)

	// We just create a new struct because that's all we really need, we satisfy the interface
	var currBatcher *ABatcher
	for _, pair := range pairList {
		// TODO: make sure that this one pointer being reused doesn't cause any issues
		if currBatcher, err = NewABatcher(maxBatchSize); err != nil {
			err = fmt.Errorf("Error creating new batcher for %s pair: %s", pair.String(), err)
			return
		}
		batchers[*pair] = currBatcher
	}

	return
}
