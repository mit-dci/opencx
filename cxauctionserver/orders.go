package cxauctionserver

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/crypto/rsw"
	"github.com/mit-dci/opencx/crypto/timelockencoders"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	logging.Infof("Got a new puzzle for auction %x", order.IntendedAuction)

	if err = s.validateEncryptedOrder(order); err != nil {
		logging.Errorf("Error validating order: %s", err)
	}

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()
	if err = s.OpencxDB.PlaceAuctionPuzzle(order); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	// This will add to the batcher
	s.OrderBatcher.AddEncrypted(order)

	return
}

// solveOrderIntoResChan solves the order puzzle and puts it in to the server's order channel.
func (s *OpencxAuctionServer) solveOrderIntoResChan(eOrder *match.EncryptedAuctionOrder) {
	var err error
	result := new(match.OrderPuzzleResult)
	result.Encrypted = eOrder

	var orderBytes []byte
	if orderBytes, err = timelockencoders.SolvePuzzleRC5(eOrder.OrderCiphertext, eOrder.OrderPuzzle); err != nil {
		result.Err = fmt.Errorf("Error solving RC5 puzzle for auction order server solve: %s", err)
		s.orderChannel <- result
		return
	}

	result.Auction = new(match.AuctionOrder)
	if err = result.Auction.Deserialize(orderBytes); err != nil {
		result.Err = fmt.Errorf("Error deserializing order from puzzle for server: %s", err)
		s.orderChannel <- result
		return
	}

	s.orderChannel <- result

	return
}

// CommitOrdersNewAuction commits to a set of decypted orders and changes the auction ID.
// TODO: figure out how to broadcast these, and where to store them, if we need to store them
func (s *OpencxAuctionServer) CommitOrdersNewAuction() (err error) {

	// Lock!
	s.dbLock.Lock()

	// First get the current auction ID
	var auctionID [32]byte
	if auctionID, err = s.CurrentAuctionID(); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Eror getting current auction id for commit: %s", err)
		return
	}

	// First, get the commitorderschannel
	var commitOrderChannel chan *match.AuctionBatch
	if commitOrderChannel, err = s.OrderBatcher.EndAuction(auctionID); err != nil {
		err = fmt.Errorf("Error ending auction while committing orders for new auction: %s", err)
		return
	}

	// Make this boi wait for the batch to come in
	go s.asyncBatchPlacer(commitOrderChannel)

	// Then get the puzzles
	var puzzles []*match.EncryptedAuctionOrder
	if puzzles, err = s.OpencxDB.ViewAuctionPuzzleBook(auctionID); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error getting auction puzzle book for commit: %s", err)
		return
	}

	sha3 := sha3.New256()
	// Add the current auction ID to be hashed
	sha3.Write(auctionID[:])
	// Then find the hash of the orders + the previous hash
	for _, pz := range puzzles {
		var pzRaw []byte
		if pzRaw, err = pz.Serialize(); err != nil {
			s.dbLock.Unlock()
			err = fmt.Errorf("Error serializing puzzle for commitment: %s", err)
			return
		}
		sha3.Write(pzRaw)
	}

	// Set the new auction ID to the hash of the orders. TODO: figure out if signing the puzzles
	// instead is a good idea, and if the dependence on the previous commitment is a good idea.
	copy(s.auctionID[:], sha3.Sum(nil))

	// Start the new auction by registering
	if err = s.OrderBatcher.RegisterAuction(s.auctionID); err != nil {
		err = fmt.Errorf("Error registering auction while committing / creating new auction: %s", err)
		return
	}

	var height uint64
	if height, err = s.OpencxDB.NewAuction(s.auctionID); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error updating auction in DB while committing orders and creating new auction: %s", err)
		return
	}

	// Unlock!
	s.dbLock.Unlock()

	logging.Infof("New height: %d", height)

	return
}

// asyncBatchPlacer waits for a batch and places it. This should be done in a goroutine
func (s *OpencxAuctionServer) asyncBatchPlacer(batchChan chan *match.AuctionBatch) {
	var err error

	defer func() {
		if err != nil {
			logging.Errorf("Error placing order asynchronously: %s", err)
		}
	}()

	batch := <-batchChan

	s.dbLock.Lock()

	var batchRes *match.BatchResult
	batchRes = s.validateBatch(batch)

	logging.Infof("Got a batch result for %x! \n\tValid orders: %d\n\tInvalid orders: %d", batchRes.OriginalBatch.AuctionID, len(batchRes.AcceptedResults), len(batchRes.RejectedResults))

	for _, acceptedOrder := range batchRes.AcceptedResults {
		if acceptedOrder.Err != nil {
			err = fmt.Errorf("Accepted order has a non-nil error: %s", acceptedOrder.Err)
			return
		}

		if err = s.OpencxDB.PlaceAuctionOrder(acceptedOrder.Auction); err != nil {
			err = fmt.Errorf("Error placing auction order with async batch placer: %s", err)
			return
		}
	}

	s.dbLock.Unlock()
	return
}

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateEncryptedOrder(order *match.EncryptedAuctionOrder) (err error) {

	var rswPuzzle *rsw.PuzzleRSW
	var ok bool
	if rswPuzzle, ok = order.OrderPuzzle.(*rsw.PuzzleRSW); !ok {
		err = fmt.Errorf("Puzzle could not be converted to RSW puzzle, invalid encrypted order")
		return
	}

	if uint64(rswPuzzle.T.Int64()) != s.t {
		err = fmt.Errorf("The time to solve the puzzle is not correct, invalid encrypted order")
		return
	}

	return
}

// validateBatch validates a batch of orders, sorting into accepted and rejected piles using validateOrder
func (s *OpencxAuctionServer) validateBatch(auctionBatch *match.AuctionBatch) (batchResult *match.BatchResult) {
	var err error

	batchResult = &match.BatchResult{
		OriginalBatch:   auctionBatch,
		RejectedResults: []*match.OrderPuzzleResult{},
		AcceptedResults: []*match.OrderPuzzleResult{},
	}

	for _, orderPzRes := range auctionBatch.Batch {
		if err = s.validateOrderResult(auctionBatch.AuctionID, orderPzRes); err != nil {
			orderPzRes.Err = fmt.Errorf("Order invalid: %s", err)
			batchResult.RejectedResults = append(batchResult.RejectedResults, orderPzRes)
		} else {
			batchResult.AcceptedResults = append(batchResult.AcceptedResults, orderPzRes)
		}
	}

	return
}

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateOrderResult(claimedAuction [32]byte, result *match.OrderPuzzleResult) (err error) {

	logging.Infof("Validating order by pubkey %x", result.Auction.Pubkey)

	if result.Err != nil {
		err = fmt.Errorf("Validation detected error early: %s", result.Err)
		return
	}

	if _, err = result.Auction.Price(); err != nil {
		err = fmt.Errorf("Orders with an indeterminable price are invalid: %s", err)
		return
	}

	if !result.Auction.IsBuySide() && !result.Auction.IsSellSide() {
		err = fmt.Errorf("Orders that aren't buy or sell side are invalid: %s", err)
		return
	}

	// We could use pub key hashes here but there might not be any reason for it
	var orderPublicKey *koblitz.PublicKey
	if orderPublicKey, err = koblitz.ParsePubKey(result.Auction.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Orders with a public key that cannot be parsed are invalid: %s", err)
		return
	}

	// e = h(asset)
	sha3 := sha3.New256()
	sha3.Write(result.Auction.SerializeSignable())
	e := sha3.Sum(nil)

	var recoveredPublickey *koblitz.PublicKey
	if recoveredPublickey, _, err = koblitz.RecoverCompact(koblitz.S256(), result.Auction.Signature, e); err != nil {
		err = fmt.Errorf("Orders whose signature cannot be verified with pubkey recovery are invalid: %s", err)
		return
	}

	if !recoveredPublickey.IsEqual(orderPublicKey) {
		err = fmt.Errorf("Recovered public key %x does not equal to pubkey %x in order", recoveredPublickey.SerializeCompressed(), orderPublicKey.SerializeCompressed())
		return
	}

	if !bytes.Equal(result.Encrypted.IntendedAuction[:], result.Auction.AuctionID[:]) {
		err = fmt.Errorf("Auction ID for decrypted and encrypted order must be equal")
		return
	}

	if !bytes.Equal(claimedAuction[:], result.Auction.AuctionID[:]) {
		err = fmt.Errorf("Auction ID must equal current auction")
		return
	}

	return
}
