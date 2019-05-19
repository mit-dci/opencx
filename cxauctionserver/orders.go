package cxauctionserver

import (
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	logging.Infof("Got a new puzzle for auction %x", order.IntendedAuction)

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()
	if err = s.OpencxDB.PlaceAuctionPuzzle(order); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	go match.SolveRC5AuctionOrderAsync(order, s.orderChannel)

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
			err = fmt.Errorf("Error serializing puzzle for commitment: %s", err)
			return
		}
		sha3.Write(pzRaw)
	}

	// Set the new auction ID to the hash of the orders. TODO: figure out if signing the puzzles
	// instead is a good idea, and if the dependence on the previous commitment is a good idea.
	copy(s.auctionID[:], sha3.Sum(nil))

	var height uint64
	if height, err = s.OpencxDB.NewAuction(s.auctionID); err != nil {
		err = fmt.Errorf("Error updating auction in DB while committing orders and creating new auction: %s", err)
		return
	}

	// Unlock!
	s.dbLock.Unlock()

	logging.Infof("Done creating new auction %x at height %d", auctionID, height)

	return
}

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateOrder(decryptedOrder *match.AuctionOrder, encryptedOrder *match.EncryptedAuctionOrder) (valid bool, err error) {

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

	valid = true
	return
}
