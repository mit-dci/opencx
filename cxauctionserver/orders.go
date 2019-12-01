package cxauctionserver

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/crypto/rsw"
	"github.com/mit-dci/opencx/crypto/timelockencoders"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrderAsync(order *match.EncryptedAuctionOrder, errChan chan error) {
	var err error
	defer func() {
		errChan <- err
	}()
	if order == nil {
		err = fmt.Errorf("Cannot place nil order, invalid")
		return
	}

	logging.Infof("Got a new puzzle for auction %x", order.IntendedAuction)

	if err = s.validateEncryptedOrder(order); err != nil {
		err = fmt.Errorf("Error validating order: %s", err)
		return
	}

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()

	// get the puzzle engine we'll use
	var pzEngine cxdb.PuzzleStore
	var ok bool
	if pzEngine, ok = s.PuzzleEngines[order.IntendedPair]; !ok {
		err = fmt.Errorf("Could not find puzzle engine for pair %s", order.IntendedPair.String())
		s.dbLock.Unlock()
		return
	}

	if err = pzEngine.PlaceAuctionPuzzle(order); err != nil {
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		s.dbLock.Unlock()
		return
	}

	var correctBatcher match.AuctionBatcher
	if correctBatcher, ok = s.OrderBatchers[order.IntendedPair]; !ok {
		err = fmt.Errorf("Could not find batcher for pair %s", order.IntendedPair.String())
		s.dbLock.Unlock()
		return
	}

	// This will add to the batcher
	if err = correctBatcher.AddEncrypted(order); err != nil {
		err = fmt.Errorf("Error adding encrypted order to batcher: %s", err)
		s.dbLock.Unlock()
		return
	}

	s.dbLock.Unlock()

	return
}

func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {
	errChan := make(chan error, 1)
	go s.PlacePuzzledOrderAsync(order, errChan)
	err = <-errChan
	close(errChan)
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

// CommitOrdersNewAuction commits to a set of encrypted orders and changes the auction ID.
// TODO: figure out how to broadcast these, and where to store them, if we need to store them
// also TODO: REWRITE because batcher is a better way of doing things
func (s *OpencxAuctionServer) CommitOrdersNewAuction(pair *match.Pair, auctionID [32]byte) (newID [32]byte, err error) {

	// Lock!
	s.dbLock.Lock()

	// get the puzzle engine we'll use
	var pzEngine cxdb.PuzzleStore
	var ok bool
	if pzEngine, ok = s.PuzzleEngines[*pair]; !ok {
		err = fmt.Errorf("Could not find puzzle engine for pair %s", pair.String())
		s.dbLock.Unlock()
		return
	}

	var matchAuctionID *match.AuctionID
	matchAuctionID = new(match.AuctionID)
	if err = matchAuctionID.UnmarshalBinary(auctionID[:]); err != nil {
		err = fmt.Errorf("Error unmarshalling auction id for CommitOrdersNewAuction: %s", err)
		s.dbLock.Unlock()
		return
	}

	var correctBatcher match.AuctionBatcher
	if correctBatcher, ok = s.OrderBatchers[*pair]; !ok {
		err = fmt.Errorf("Could not find batcher for pair %s", pair.String())
		s.dbLock.Unlock()
		return
	}

	// First, get the commitorderschannel
	var commitOrderChannel chan *match.AuctionBatch
	if commitOrderChannel, err = correctBatcher.EndAuction(auctionID); err != nil {
		err = fmt.Errorf("Error ending auction while committing orders for new auction: %s", err)
		s.dbLock.Unlock()
		return
	}

	// Make this boi wait for the batch to come in
	go s.asyncBatchPlacer(commitOrderChannel)

	// Then get the puzzles
	var puzzles []*match.EncryptedAuctionOrder
	if puzzles, err = pzEngine.ViewAuctionPuzzleBook(matchAuctionID); err != nil {
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
			s.dbLock.Unlock()
			return
		}
		sha3.Write(pzRaw)
	}

	// Set the new auction ID to the hash of the orders. TODO: figure out if
	// dependence on the previous commitment is a good idea.
	var newAuctionID [32]byte
	copy(newAuctionID[:], sha3.Sum(nil))
	// TODO: sign
	// TODO: how to broadcast and timestamp these?

	// Start the new auction by registering
	if err = correctBatcher.RegisterAuction(newAuctionID); err != nil {
		err = fmt.Errorf("Error registering auction while committing / creating new auction: %s", err)
		s.dbLock.Unlock()
		return
	}

	// var height uint64
	// if height, err = s.MatchingEngine.NewAuctionHeight(newAuctionID); err != nil {
	// 	err = fmt.Errorf("Error updating auction in DB while committing orders and creating new auction: %s", err)
	// 	s.dbLock.Unlock()
	// 	return
	// }

	// Unlock!
	s.dbLock.Unlock()
	newID = newAuctionID

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

	// TODO: fix this spaghet
	batch := <-batchChan
	batchChan <- batch

	s.dbLock.Lock()

	var batchRes *match.BatchResult
	batchRes = s.validateBatch(batch)

	logging.Infof("Got a batch result for %x! \n\tValid orders: %d\n\tInvalid orders: %d", batchRes.OriginalBatch, len(batchRes.AcceptedResults), len(batchRes.RejectedResults))

	for _, acceptedOrder := range batchRes.AcceptedResults {
		if acceptedOrder.Err != nil {
			err = fmt.Errorf("Accepted order has a non-nil error: %s", acceptedOrder.Err)
			s.dbLock.Unlock()
			return
		}

		// TODO work out auction placing
		// if err = s.OpencxDB.PlaceAuctionOrder(acceptedOrder.Auction); err != nil {
		// 	err = fmt.Errorf("Error placing auction order with async batch placer: %s", err)
		// 	return
		// }
	}

	s.dbLock.Unlock()
	return
}

func (s *OpencxAuctionServer) PlaceBatch(batch *match.AuctionBatch) (err error) {

	s.dbLock.Lock()

	var auctionEngine match.AuctionEngine
	var ok bool

	var batchRes *match.BatchResult = s.validateBatch(batch)

	logging.Infof("Got a batch result for %x! \n\tValid orders: %d\n\tInvalid orders: %d", batchRes.OriginalBatch, len(batchRes.AcceptedResults), len(batchRes.RejectedResults))

	var auctionIDList map[match.AuctionID]bool = make(map[match.AuctionID]bool)
	for _, acceptedOrder := range batchRes.AcceptedResults {
		if acceptedOrder.Err != nil {
			err = fmt.Errorf("Accepted order has a non-nil error: %s", acceptedOrder.Err)
			s.dbLock.Unlock()
			return
		}

		if auctionEngine, ok = s.MatchingEngines[acceptedOrder.Auction.TradingPair]; !ok {
			err = fmt.Errorf("Could not find matching engine for pair %s", acceptedOrder.Auction.TradingPair.String())
			s.dbLock.Unlock()
			return
		}

		var idStruct *match.AuctionID = new(match.AuctionID)
		if err = idStruct.UnmarshalBinary(acceptedOrder.Auction.AuctionID[:]); err != nil {
			err = fmt.Errorf("Error unmarshalling auction ID: %s", err)
			s.dbLock.Unlock()
			return
		}

		auctionIDList[*idStruct] = true

		var placeRes *match.AuctionOrderIDPair
		if placeRes, err = auctionEngine.PlaceAuctionOrder(acceptedOrder.Auction, idStruct); err != nil {
			err = fmt.Errorf("Error placing auction order with async batch placer: %s", err)
			s.dbLock.Unlock()
			return
		}

		logging.Infof("Placed order %x for auction %x", placeRes.OrderID[:], acceptedOrder.Auction.AuctionID)

	}

	// Now we're going to match it
	var currIDPtr *match.AuctionID
	for id, _ := range auctionIDList {
		// I don't want to reuse the `id` loop var pointer
		currIDPtr = new(match.AuctionID)
		*currIDPtr = id
		// We ignore the Order executions because we're not doing anything about them yet
		if _, _, err = auctionEngine.MatchAuctionOrders(currIDPtr); err != nil {
			err = fmt.Errorf("Error matching orders for PlaceBatch: %s", err)
			s.dbLock.Unlock()
			return
		}
	}
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

		var pr float64
		if pr, err = orderPzRes.Auction.Price(); err != nil {
			orderPzRes.Err = fmt.Errorf("Error getting price from order: %s", err)
		}
		// TODO: this is to protect the database, this is why switching to a better price system would be a good idea
		if pr > float64(10000000000000000000000) {
			orderPzRes.Err = fmt.Errorf("Price too high, complain online if you want the maximum price increased, or lower your price")
		}
		if pr < float64(1)/float64(1000000) {
			orderPzRes.Err = fmt.Errorf("Price too low, complain online if you want the minimum price decreased, or increase your price")
		}
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

	if result == nil {
		err = fmt.Errorf("Result cannot be nil, please enter valid input")
		return
	}

	if result.Auction == nil {
		err = fmt.Errorf("Auction in result cannot be nil, please enter valid input")
		return
	}

	if result.Encrypted == nil {
		err = fmt.Errorf("Encrypted order in result cannot be nil, please enter valid input")
		return
	}
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

func (s *OpencxAuctionServer) runMatching(auctionID *match.AuctionID, pair *match.Pair) (err error) {

	s.dbLock.Lock()

	var matchEngine match.AuctionEngine
	var ok bool
	if matchEngine, ok = s.MatchingEngines[*pair]; !ok {
		err = fmt.Errorf("Error getting correct matching engine for pair %s for runMatching", pair)
		s.dbLock.Unlock()
		return
	}

	// We can now calculate a clearing price and run the matching algorithm
	var orderExecs []*match.OrderExecution
	var setExecs []*match.SettlementExecution
	if orderExecs, setExecs, err = matchEngine.MatchAuctionOrders(auctionID); err != nil {
		err = fmt.Errorf("Error matching orders for running matching: %s", err)
		s.dbLock.Unlock()
		return
	}

	var orderbook match.AuctionOrderbook
	if orderbook, ok = s.Orderbooks[*pair]; !ok {
		err = fmt.Errorf("Error getting correct orderbook for pair %s for runMatching", pair)
		s.dbLock.Unlock()
		return
	}

	for _, orderExec := range orderExecs {
		if err = orderbook.UpdateBookExec(orderExec); err != nil {
			err = fmt.Errorf("Error updating book for order execution: %s", err)
			s.dbLock.Unlock()
			return
		}
	}

	var setCoinParam *coinparam.Params
	var setEngine match.SettlementEngine
	for _, settlementExec := range setExecs {
		// get coin param for debited
		if setCoinParam, err = settlementExec.Asset.CoinParamFromAsset(); err != nil {
			err = fmt.Errorf("Error getting coin param for runMatching: %s", err)
			s.dbLock.Unlock()
			return
		}

		if setEngine, ok = s.SettlementEngines[setCoinParam]; !ok {
			err = fmt.Errorf("Error getting correct settlement engine for runMatching")
			s.dbLock.Unlock()
			return
		}

		var settleValid bool
		if settleValid, err = setEngine.CheckValid(settlementExec); err != nil {
			err = fmt.Errorf("Error checking valid for runMatching: %s", err)
			return
		}

		if settleValid {

			if _, err = setEngine.ApplySettlementExecution(settlementExec); err != nil {
				err = fmt.Errorf("Error applyin settlement execution for runMatching: %s", err)
				s.dbLock.Unlock()
				return
			}

		} else {
			err = fmt.Errorf("Settlement invalid for some reason, maybe run matching again to see if anything changes")
			return
		}

	}

	s.dbLock.Unlock()

	return
}
