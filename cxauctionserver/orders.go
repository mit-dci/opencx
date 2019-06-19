package cxauctionserver

import (
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/crypto/rsw"
	"github.com/mit-dci/opencx/crypto/timelockencoders"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	logging.Infof("Got a new puzzle for auction %x", order.IntendedAuction)

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()
	if err = s.PuzzleEngine.PlaceAuctionPuzzle(order); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	if err = s.validateEncryptedOrder(order); err != nil {
		logging.Errorf("Error validating order: %s", err)
	}

	go s.solveOrderIntoResChan(order)

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

	// Then get the puzzles
	var puzzles []*match.EncryptedAuctionOrder
	if puzzles, err = s.PuzzleEngine.ViewAuctionPuzzleBook(auctionID); err != nil {
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
			err = fmt.Errorf("Error serializing puzzle for commitment: %s", err)
			return
		}
		sha3.Write(pzRaw)
	}

	// Set the new auction ID to the hash of the orders. TODO: figure out if signing the puzzles
	// instead is a good idea, and if the dependence on the previous commitment is a good idea.
	copy(s.auctionID[:], sha3.Sum(nil))

	// TODO: abstract out height storage, in batcher
	// var height uint64
	// if height, err = s.MatchingEngine.NewAuctionHeight(s.auctionID); err != nil {
	// 	err = fmt.Errorf("Error updating auction in DB while committing orders and creating new auction: %s", err)
	// 	return
	// }

	// Unlock!
	s.dbLock.Unlock()

	// logging.Infof("Done creating new auction %x at height %d", auctionID, height)

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

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateOrder(decryptedOrder *match.AuctionOrder, encryptedOrder *match.EncryptedAuctionOrder) (err error) {

	logging.Infof("Validating order by pubkey %x", decryptedOrder.Pubkey)

	if _, err = decryptedOrder.Price(); err != nil {
		err = fmt.Errorf("Orders with an indeterminable price are invalid: %s", err)
		return
	}

	if !decryptedOrder.IsBuySide() && !decryptedOrder.IsSellSide() {
		err = fmt.Errorf("Orders that aren't buy or sell side are invalid: %s", err)
		return
	}

	// We could use pub key hashes here but there might not be any reason for it
	var orderPublicKey *koblitz.PublicKey
	if orderPublicKey, err = koblitz.ParsePubKey(decryptedOrder.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Orders with a public key that cannot be parsed are invalid: %s", err)
		return
	}

	// e = h(asset)
	sha3 := sha3.New256()
	sha3.Write(decryptedOrder.SerializeSignable())
	e := sha3.Sum(nil)

	var recoveredPublickey *koblitz.PublicKey
	if recoveredPublickey, _, err = koblitz.RecoverCompact(koblitz.S256(), decryptedOrder.Signature, e); err != nil {
		err = fmt.Errorf("Orders whose signature cannot be verified with pubkey recovery are invalid: %s", err)
		return
	}

	if !recoveredPublickey.IsEqual(orderPublicKey) {
		err = fmt.Errorf("Recovered public key %x does not equal to pubkey %x in order", recoveredPublickey.SerializeCompressed(), orderPublicKey.SerializeCompressed())
		return
	}

	// TODO: figure out how to deal with auctionID
	// if !bytes.Equal(s.auctionID[:], decryptedOrder.AuctionID[:]) {
	// 	err = fmt.Errorf("Auction ID must equal current auction")
	// 	return
	// }

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
