package hashtimelock

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"hash"

	"github.com/mit-dci/opencx/crypto"
)

// HashTimelock is a struct that holds all data necessary to implement a timelock
type HashTimelock struct {
	// timelockSeed is the initial data that then gets hashed by a hash function
	timelockSeed []byte
	// hashFunction is the hash function to be used by this hash timelock
	hashFunction hash.Hash
	// timeToRun is the amount of iterations needed to run
	timeToRun uint64
}

func (ht *HashTimelock) setupHashPuzzle(seed []byte, hashFunction hash.Hash) {
	ht.timelockSeed = seed
	ht.hashFunction = hashFunction
	return
}

// New creates a new hash timelock with seed bytes and a hash function
func New(seed []byte, hashFunction hash.Hash) (hashTimelock crypto.Timelock) {
	ht := &HashTimelock{}
	ht.setupHashPuzzle(seed, hashFunction)
	hashTimelock = ht
	return
}

// SetupTimelockPuzzle sends key k to the future in time t, returning a puzzle based on sequential hashing and an answer
func (ht *HashTimelock) SetupTimelockPuzzle(t uint64) (puzzle crypto.Puzzle, answer []byte, err error) {
	ht.timeToRun = t
	answer = make([]byte, ht.hashFunction.Size())

	copy(answer[:], ht.timelockSeed)

	for i := uint64(0); i < ht.timeToRun; i++ {
		ht.hashFunction.Reset()
		ht.hashFunction.Write(answer[:])
		copy(answer[:], ht.hashFunction.Sum(nil))
	}
	// hash time lock puzzles are their own timelocks as well as puzzles
	puzzle = ht
	return
}

// Solve solves the hash puzzle and returns the answer, or fails
func (ht *HashTimelock) Solve() (answer []byte, err error) {
	answer = make([]byte, ht.hashFunction.Size())
	copy(answer[:], ht.timelockSeed)
	for i := uint64(0); i < ht.timeToRun; i++ {
		ht.hashFunction.Reset()
		ht.hashFunction.Write(answer[:])
		copy(answer[:], ht.hashFunction.Sum(nil))
	}
	ht.hashFunction.Reset()
	return
}

// Serialize turns the hash timelock puzzle into something that can be sent over the wire
func (ht *HashTimelock) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register hashTimelock interface
	gob.Register(HashTimelock{})

	// create a new encoder writing to the buffer
	enc := gob.NewEncoder(&b)

	// encode the puzzle in the buffer
	if err = enc.Encode(ht); err != nil {
		err = fmt.Errorf("Error encoding puzzle: %s", err)
		return
	}

	// Get the bytes from the buffer
	raw = b.Bytes()

	return
}

// Deserialize turns the hash timelock puzzle into something that can be sent over the wire
func (ht *HashTimelock) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register hashTimelock interface
	gob.Register(HashTimelock{})

	// create a new encoder writing to the buffer
	dec := gob.NewDecoder(b)

	// encode the puzzle in the buffer
	if err = dec.Decode(ht); err != nil {
		err = fmt.Errorf("Error encoding puzzle: %s", err)
		return
	}

	return
}
