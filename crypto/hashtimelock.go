package crypto

import "hash"

// HashTimelock is a struct that holds all data necessary to implement a timelock
type HashTimelock struct {
	// timelockSeed is the initial data that then gets hashed by a hash function
	TimelockSeed []byte
	// hashFunction is the hash function to be used by this hash timelock
	HashFunction hash.Hash
}

// New creates a new hash timelock
func New() (ht *Puzzle) {
	return
}

// SetupTimelock sends key k to the future in time t, returning a puzzle based on sequential hashing and an answer
func (ht *HashTimelock) SetupTimelock(t uint64) (puzzle *Puzzle, answer []byte, err error) {
	return
}

// Solve solves the hash puzzle and returns the answer, or fails
func (ht *HashTimelock) Solve() (answer []byte, err error) {
	return
}
