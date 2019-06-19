package match

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/mit-dci/opencx/crypto"
	"github.com/mit-dci/opencx/crypto/hashtimelock"
	"github.com/mit-dci/opencx/crypto/rsw"
	"github.com/mit-dci/opencx/crypto/timelockencoders"
)

// EncryptedAuctionOrder represents an encrypted Auction Order, so a ciphertext and a puzzle whos solution is a key, and an intended auction.
type EncryptedAuctionOrder struct {
	OrderCiphertext []byte
	OrderPuzzle     crypto.Puzzle
	IntendedAuction [32]byte
	IntendedPair    Pair
}

// SolveRC5AuctionOrderAsync solves order puzzles and creates auction orders from them. This should be run in a goroutine.
func SolveRC5AuctionOrderAsync(e *EncryptedAuctionOrder, puzzleResChan chan *OrderPuzzleResult) {
	var err error
	result := new(OrderPuzzleResult)
	result.Encrypted = e

	var orderBytes []byte
	if orderBytes, err = timelockencoders.SolvePuzzleRC5(e.OrderCiphertext, e.OrderPuzzle); err != nil {
		result.Err = fmt.Errorf("Error solving RC5 puzzle for auction order: %s", err)
		puzzleResChan <- result
		return
	}

	result.Auction = new(AuctionOrder)
	if err = result.Auction.Deserialize(orderBytes); err != nil {
		result.Err = fmt.Errorf("Error deserializing order gotten from puzzle: %s", err)
		puzzleResChan <- result
		return
	}

	puzzleResChan <- result

	return
}

// Serialize serializes the encrypted order using gob
func (e *EncryptedAuctionOrder) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register the rsw puzzle and hashtimelock puzzle
	gob.Register(new(rsw.PuzzleRSW))

	// register the hashtimelock (puzzle and timelock are same thing)
	gob.Register(new(hashtimelock.HashTimelock))

	// register the pair
	gob.Register(new(Pair))

	// register the puzzle interface
	gob.RegisterName("puzzle", new(crypto.Puzzle))

	// register the encrypted auction order interface with gob
	gob.RegisterName("order", new(EncryptedAuctionOrder))

	// create a new encoder writing to our buffer
	enc := gob.NewEncoder(&b)

	// encode the encrypted auction order in the buffer
	if err = enc.Encode(e); err != nil {
		err = fmt.Errorf("Error encoding encrypted auction order :%s", err)
		return
	}

	// Get the bytes finally
	raw = b.Bytes()

	return
}

// Deserialize deserializes the raw bytes into the encrypted auction order receiver
func (e *EncryptedAuctionOrder) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register the rsw puzzle and hashtimelock puzzle
	gob.Register(new(rsw.PuzzleRSW))

	// register the hashtimelock (puzzle and timelock are same thing)
	gob.Register(new(hashtimelock.HashTimelock))

	// register the pair
	gob.Register(new(Pair))

	// register the puzzle interface
	gob.RegisterName("puzzle", new(crypto.Puzzle))

	// register the encrypted auction order interface with gob
	gob.RegisterName("order", new(EncryptedAuctionOrder))

	// create a new decoder writing to the buffer
	dec := gob.NewDecoder(b)

	// decode the encrypted auction order in the buffer
	if err = dec.Decode(e); err != nil {
		err = fmt.Errorf("Error decoding encrypted auction order: %s", err)
		return
	}

	return
}
