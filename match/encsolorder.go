package match

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/mit-dci/opencx/crypto/rsw"
)

// EncryptedSolutionOrder represents an encrypted Solution Order, so a
// ciphertext and a puzzle solution that is a key, and an intended auction.
type EncryptedSolutionOrder struct {
	OrderCiphertext []byte        `json:"orderciphertext"`
	OrderPuzzle     rsw.PuzzleRSW `json:"orderpuzzle"`
	IntendedAuction AuctionID     `json:"intendedauction"`
	IntendedPair    Pair          `json:"intendedpair"`
}

// Serialize uses gob encoding to turn the encrypted solution order
// into bytes.
func (es *EncryptedSolutionOrder) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register SolutionOrder interface
	gob.Register(EncryptedSolutionOrder{})

	// create a new encoder writing to the buffer
	enc := gob.NewEncoder(&b)

	// encode the puzzle in the buffer
	if err = enc.Encode(es); err != nil {
		err = fmt.Errorf("Error encoding encryptedsolutionorder: %s", err)
		return
	}

	// Get the bytes from the buffer
	raw = b.Bytes()
	return
}

// Deserialize turns the encrypted solution order from bytes into a
// usable struct.
func (es *EncryptedSolutionOrder) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register SolutionOrder
	gob.Register(EncryptedSolutionOrder{})

	// create a new decoder writing to the buffer
	dec := gob.NewDecoder(b)

	// decode the solutionorder in the buffer
	if err = dec.Decode(es); err != nil {
		err = fmt.Errorf("Error decoding encryptedsolutionorder: %s", err)
		return
	}

	return
}
