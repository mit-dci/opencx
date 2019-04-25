package hashtimelock

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"github.com/btcsuite/fastsha256"
	"github.com/dchest/siphash"
	"github.com/minio/highwayhash"
	"golang.org/x/crypto/blake2b"
)

// TestZeroTimeSHA256 is a very useful test
func TestZeroTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxzero"))
	hashFunction := sha256.New()
	// eh this whole New thing looks bad but really it'll look like hashtimelock.New(seed, hfunc) anyways which is a really nice way to do it so this will be good actually
	hashPuzzle := New(seed, hashFunction)
	time := uint64(0)
	// throw away the answer because we already know what it's supposed to be: the seed
	puzzle, _, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, seed) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, seed, puzzleAns)
	}
}

// TestOneTimeSHA256 is also a useful test
func TestOneTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxone"))
	hashFunction := sha256.New()
	// eh this whole New thing looks bad but really it'll look like hashtimelock.New(seed, hfunc) anyways which is a really nice way to do it so this will be good actually
	hashPuzzle := New(seed, hashFunction)
	time := uint64(1)
	// throw away the answer because we already know what it's supposed to be: the seed
	puzzle, _, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}

	// hash once
	testSHA := sha256.New()

	// just hash once
	if _, err = testSHA.Write(seed); err != nil {
		t.Fatalf("Could not write to hash function: %s\n", err)
	}
	seedSum := testSHA.Sum(nil)

	if !bytes.Equal(puzzleAns, seedSum) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, seedSum, puzzleAns)
	}
}

// if Zero and One work then 10 will work because it's just changing the number of times the hash is done.
// blah blah blah induction, I don't want to have to copy paste loops that are basically the same algorithm as what's in the setup / solve
func TestTenTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxten"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(10)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

func TestHundredTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundred"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

func TestThousandTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxthousand"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(1000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}
func TestTenThousandTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxtenthousand"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(10000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

func TestHundredThousandTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredthousand"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// TestMillionTimeSHA256 is really fun, but not the most fun
func TestMillionTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxmillion"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(1000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

func TestTenMillionTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxtenmillion"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(10000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

func TestHundredMillionTimeSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredmillion"))
	hashFunction := sha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// Doing this to see how much a better algo makes a difference (figure out if the bottleneck is Read/Write speed)
func TestHundredMillionTimeFastSHA256(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredmillion"))
	hashFunction := fastsha256.New()
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// lol it's slower than crypto/sha256???

// Blake2B is cool
func TestHundredMillionTimeBlake2B(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredmillion"))
	hashFunction, err := blake2b.New256(nil)
	if err != nil {
		t.Fatalf("Could not set up blake2b: %s\n", err)
	}
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// Now let's see, siphash is supposed to be fast
func TestHundredMillionTimeSipHash(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredmillion"))
	hashFunction := siphash.New(seed)
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// Now let's see, highwayhash is supposed to be faster
func TestHundredMillionTimeHighwayHash(t *testing.T) {
	seed := make([]byte, 32)
	copy(seed[:], []byte("opencxhundredmillion"))
	hashFunction, err := highwayhash.New(seed)
	if err != nil {
		t.Fatalf("Could not create highwayhash: %s\n", err)
	}
	hashPuzzle := New(seed, hashFunction)
	time := uint64(100000000)
	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
	if err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		t.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		t.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
}

// TestBillionTimeSHA256 is even more fun, but we can't run it because we need the time to be > 10 minutes
// func TestBillionTimeSHA256(t *testing.T) {
// seed := make([]byte, 32)
// copy(seed[:], []byte("opencxbillion"))
// 	hashFunction := sha256.New()
// 	hashPuzzle := New(seed, hashFunction)
// 	time := uint64(1000000000)

// 	// so we solve it once here, this can take some time
// 	puzzle, expectedAns, err := hashPuzzle.SetupTimelockPuzzle(time)
// 	if err != nil {
// 		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
// 	}

// 	// now we do the exact same thing all over again lol
// 	puzzleAns, err := puzzle.Solve()
// 	if err != nil {
// 		t.Fatalf("Error solving puzzle: %s\n", err)
// 	}
// 	if !bytes.Equal(puzzleAns, expectedAns) {
// 		t.Fatalf("Answer did not equal puzzle for time = 0. Expected %x, got %x\n", expectedAns, puzzleAns)
// 	}
// }
