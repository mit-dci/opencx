package match

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"fmt"
	"math/big"

	"github.com/mit-dci/opencx/crypto/timelockencoders"
)

// SolutionOrder is an order and modulus that are together.
// This includes an order, and the puzzle modulus factors.
type SolutionOrder struct {
	userOrder AuctionOrder `json:"userorder"`
	p         *big.Int     `json:"p"`
	q         *big.Int     `json:"q"`
}

// NewSolutionOrderFromOrder creates a new SolutionOrder from an
// already existing AuctionOrder, with a specified number of bits for
// an rsa key.
func NewSolutionOrderFromOrder(aucOrder *AuctionOrder, rsaKeyBits uint64) (solOrder SolutionOrder, err error) {
	if aucOrder == nil {
		err = fmt.Errorf("Cannot create solution order from nil auction order")
		return
	}

	rsaKeyBitsInt := int(rsaKeyBits)

	// generate primes p and q
	var rsaPrivKey *rsa.PrivateKey
	if rsaPrivKey, err = rsa.GenerateMultiPrimeKey(rand.Reader, 2, rsaKeyBitsInt); err != nil {
		err = fmt.Errorf("Could not generate primes for RSA: %s", err)
		return
	}
	if len(rsaPrivKey.Primes) != 2 {
		err = fmt.Errorf("For some reason the RSA Privkey has != 2 primes, this should not be the case for RSW, we only need p and q")
		return
	}

	// finally set p, q, and the auction order.
	solOrder.p = new(big.Int).SetBytes(rsaPrivKey.Primes[0].Bytes())
	solOrder.q = new(big.Int).SetBytes(rsaPrivKey.Primes[1].Bytes())
	solOrder.userOrder = *aucOrder
	return
}

// EncryptSolutionOrder encrypts a solution order and creates a puzzle
// along with the encrypted order
func (so *SolutionOrder) EncryptSolutionOrder(t uint64) (encSolOrder EncryptedSolutionOrder, err error) {
	// Try serializing the solution order
	var rawSolOrder []byte
	if rawSolOrder, err = so.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing solution order for encryption: %s", err)
		return
	}

	if encSolOrder.OrderCiphertext, encSolOrder.OrderPuzzle, err = timelockencoders.CreateRC5RSWPuzzleWithPrimes(uint64(2), t, rawSolOrder, so.p, so.q); err != nil {
		err = fmt.Errorf("Error creating puzzle from auction order: %s", err)
		return
	}

	// make sure they match
	encSolOrder.IntendedAuction = so.userOrder.AuctionID
	encSolOrder.IntendedPair = so.userOrder.TradingPair
	return
}

// Serialize uses gob encoding to turn the solution order into bytes.
func (so *SolutionOrder) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register SolutionOrder interface
	gob.Register(SolutionOrder{})

	// create a new encoder writing to the buffer
	enc := gob.NewEncoder(&b)

	// encode the puzzle in the buffer
	if err = enc.Encode(so); err != nil {
		err = fmt.Errorf("Error encoding solutionorder: %s", err)
		return
	}

	// Get the bytes from the buffer
	raw = b.Bytes()
	return
}

// Deserialize turns the solution order from bytes into a usable
// struct.
func (so *SolutionOrder) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register SolutionOrder
	gob.Register(SolutionOrder{})

	// create a new decoder writing to the buffer
	dec := gob.NewDecoder(b)

	// decode the solutionorder in the buffer
	if err = dec.Decode(so); err != nil {
		err = fmt.Errorf("Error decoding solutionorder: %s", err)
		return
	}

	return
}
