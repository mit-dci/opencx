package rsw

import (
	"bytes"
	"fmt"
	"testing"
)

func createTest2048A2(time uint64, t *testing.T) {
	key := make([]byte, 32)
	copy(key[:], []byte(fmt.Sprintf("opencx%d", time)))
	rswTimelock, err := New2048A2(key)
	if err != nil {
		t.Fatalf("There was an error creating a new timelock puzzle: %s", err)
	}
	puzzle, expectedAns, err := rswTimelock.SetupTimelockPuzzle(time)
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
	return
}
func TestZero2048A2(t *testing.T) {
	createTest2048A2(0, t)
	return
}

func TestOne2048A2(t *testing.T) {
	createTest2048A2(1, t)
	return
}

func TestTen2048A2(t *testing.T) {
	createTest2048A2(10, t)
	return
}

func TestHundred2048A2(t *testing.T) {
	createTest2048A2(100, t)
	return
}

func TestThousand2048A2(t *testing.T) {
	createTest2048A2(1000, t)
	return
}

func TestTenThousand2048A2(t *testing.T) {
	createTest2048A2(10000, t)
	return
}

func TestHundredThousand2048A2(t *testing.T) {
	createTest2048A2(100000, t)
	return
}
