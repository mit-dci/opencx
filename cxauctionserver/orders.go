package cxauctionserver

import (
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

// PlacePuzzledOrder places a timelock encrypted order. It also starts to decrypt the order in a goroutine.
func (s *OpencxAuctionServer) PlacePuzzledOrder(order *match.EncryptedAuctionOrder) (err error) {

	// Placing an auction puzzle is how the exchange will then recall and commit to a set of puzzles.
	s.dbLock.Lock()
	if err = s.OpencxDB.PlaceAuctionPuzzle(order); err != nil {
		s.dbLock.Unlock()
		err = fmt.Errorf("Error placing puzzled order: \n%s", err)
		return
	}
	s.dbLock.Unlock()

	// send order solving to channel
	go order.SolveRC5AuctionOrderAsync(s.orderChannel)

	return
}

// validateOrder is how the server checks that an order is valid, and checks out with its corresponding encrypted order
func (s *OpencxAuctionServer) validateOrder(decryptedOrder *match.AuctionOrder, encryptedOrder *match.EncryptedAuctionOrder) (valid bool, err error) {

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
		err = fmt.Errorf("Recovered public key does not equal to pubkey in order")
		return
	}

	valid = true
	return
}
