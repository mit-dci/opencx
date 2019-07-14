package rsw

import (
	"bytes"
	"fmt"
	"log"
	"runtime"
	"testing"
)

// This is how you create a solvable RSW timelock puzzle.
func ExampleTimelockRSW_SetupTimelockPuzzle() {
	// Allocate memory for key
	key := make([]byte, 32)
	// set key to be some bytes that we want to be the solution to the puzzle
	copy(key[:], []byte(fmt.Sprint("!!! secret < 32 bytes !!!")))
	// Create a new timelock. A timelock can be used to create puzzles for a key with a certain time.
	rswTimelock, err := New2048A2(key)
	if err != nil {
		log.Fatalf("Error creating a new timelock puzzle: %s", err)
	}

	// set the time to be some big number
	t := uint64(1000000)

	// Create the puzzle. Puzzles can be solved.
	puzzle, expectedAns, err := rswTimelock.SetupTimelockPuzzle(t)
	if err != nil {
		log.Fatalf("Error creating puzzle: %s", err)
	}

	fmt.Printf("Puzzle nil? %t. Expected Answer: %s", puzzle == nil, string(expectedAns))
	// Puzzle nil? false. Expected Answer: !!! secret < 32 bytes !!!

}

func createSolveConcurrentN(time uint64, n int, t *testing.T) {
	doneChan := make(chan bool, n)
	for i := 0; i < n; i++ {
		go createSolveTest2048A2Async(time, doneChan, t)
	}

	// Wait for our things to return - there may be a better way to do this with select
	for i := 0; i < n; i++ {
		<-doneChan
	}

	return
}

func createSolveConcurrentNBench(time uint64, n int, b *testing.B) {
	doneChan := make(chan bool, n)
	for i := 0; i < n; i++ {
		go createSolveBench2048A2Async(time, doneChan, b)
	}

	// Wait for our things to return - there may be a better way to do this with select
	for i := 0; i < n; i++ {
		<-doneChan
	}

	return
}

func createSolveConcurrent(time uint64, t *testing.T) {
	createSolveConcurrentN(time, runtime.NumCPU(), t)
	return
}

func createSolveTest2048A2Async(time uint64, doneChan chan bool, t *testing.T) {
	key := make([]byte, 32)
	copy(key[:], []byte(fmt.Sprintf("opencxcreatesolve%d", time)))
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

	doneChan <- true
	return
}

func createSolveBench2048A2Async(time uint64, doneChan chan bool, b *testing.B) {
	key := make([]byte, 32)
	copy(key[:], []byte(fmt.Sprintf("opencxcreatesolve%d", time)))
	rswTimelock, err := New2048A2(key)
	if err != nil {
		b.Fatalf("There was an error creating a new timelock puzzle: %s", err)
	}
	puzzle, expectedAns, err := rswTimelock.SetupTimelockPuzzle(time)
	if err != nil {
		b.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		b.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		b.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}

	doneChan <- true
	return
}

func createSolveBench2048A2(time uint64, b *testing.B) {
	key := make([]byte, 32)
	copy(key[:], []byte(fmt.Sprintf("opencxcreatesolve%d", time)))
	rswTimelock, err := New2048A2(key)
	if err != nil {
		b.Fatalf("There was an error creating a new timelock puzzle: %s", err)
	}
	puzzle, expectedAns, err := rswTimelock.SetupTimelockPuzzle(time)
	if err != nil {
		b.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	puzzleAns, err := puzzle.Solve()
	if err != nil {
		b.Fatalf("Error solving puzzle: %s\n", err)
	}
	if !bytes.Equal(puzzleAns, expectedAns) {
		b.Fatalf("Answer did not equal puzzle for time = %d. Expected %x, got %x\n", time, expectedAns, puzzleAns)
	}
	return
}

func createSolveTest2048A2(time uint64, t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte(fmt.Sprintf("opencxcreatesolve%d", time)))
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

func createTest2048A2(time uint64, t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte(fmt.Sprintf("opencx%d", time)))
	rswTimelock, err := New2048A2(key)
	if err != nil {
		t.Fatalf("There was an error creating a new timelock puzzle: %s", err)
	}
	if _, _, err = rswTimelock.SetupTimelockPuzzle(time); err != nil {
		t.Fatalf("There was an error setting up the timelock puzzle: %s\n", err)
	}
	return
}

// TestCreate tests are to show that it's not the creation/setup steps that we're waiting for, it's the solve step.
// Solving whatever is created by TestCreateQuintrillion2048A2 is going to take a long time, much past our 10 minute
// testing limit
func TestCreateZero2048A2(t *testing.T) {
	createTest2048A2(0, t)
	return
}

func TestCreateOne2048A2(t *testing.T) {
	createTest2048A2(1, t)
	return
}

func TestCreateTen2048A2(t *testing.T) {
	createTest2048A2(10, t)
	return
}

func TestCreateHundred2048A2(t *testing.T) {
	createTest2048A2(100, t)
	return
}

func TestCreateThousand2048A2(t *testing.T) {
	createTest2048A2(1000, t)
	return
}

func TestCreateTenThousand2048A2(t *testing.T) {
	createTest2048A2(10000, t)
	return
}

func TestCreateHundredThousand2048A2(t *testing.T) {
	createTest2048A2(100000, t)
	return
}

func TestCreateMillion2048A2(t *testing.T) {
	createTest2048A2(1000000, t)
	return
}

func TestCreateTenMillion2048A2(t *testing.T) {
	createTest2048A2(10000000, t)
	return
}

func TestCreateHundredMillion2048A2(t *testing.T) {
	createTest2048A2(100000000, t)
	return
}

func TestCreateBillion2048A2(t *testing.T) {
	createTest2048A2(1000000000, t)
	return
}

func TestCreateTenBillion2048A2(t *testing.T) {
	createTest2048A2(10000000000, t)
	return
}

func TestCreateHundredBillion2048A2(t *testing.T) {
	createTest2048A2(100000000000, t)
	return
}

func TestCreateTrillion2048A2(t *testing.T) {
	createTest2048A2(1000000000000, t)
	return
}

func TestCreateTenTrillion2048A2(t *testing.T) {
	createTest2048A2(10000000000000, t)
	return
}

func TestCreateHundredTrillion2048A2(t *testing.T) {
	createTest2048A2(100000000000000, t)
	return
}

func TestCreateQuadrillion2048A2(t *testing.T) {
	createTest2048A2(1000000000000000, t)
	return
}

func TestCreateTenQuadrillion2048A2(t *testing.T) {
	createTest2048A2(10000000000000000, t)
	return
}

func TestCreateHundredQuadrillion2048A2(t *testing.T) {
	createTest2048A2(100000000000000000, t)
	return
}

func TestCreateQuintrillion2048A2(t *testing.T) {
	createTest2048A2(1000000000000000000, t)
	return
}

func TestCreateLCSTime2048A2(t *testing.T) {
	createTest2048A2(79685186856218, t)
	return
}

func TestZero2048A2(t *testing.T) {
	createSolveTest2048A2(0, t)
	return
}

func TestOne2048A2(t *testing.T) {
	createSolveTest2048A2(1, t)
	return
}

func TestTen2048A2(t *testing.T) {
	createSolveTest2048A2(10, t)
	return
}

func TestHundred2048A2(t *testing.T) {
	createSolveTest2048A2(100, t)
	return
}

func TestThousand2048A2(t *testing.T) {
	createSolveTest2048A2(1000, t)
	return
}

func TestTenThousand2048A2(t *testing.T) {
	createSolveTest2048A2(10000, t)
	return
}

func TestHundredThousand2048A2(t *testing.T) {
	createSolveTest2048A2(100000, t)
	return
}

func TestMillion2048A2(t *testing.T) {
	createSolveTest2048A2(1000000, t)
	return
}

func TestTenMillion2048A2(t *testing.T) {
	createSolveTest2048A2(10000000, t)
	return
}

func BenchmarkHundredMillion2048A2(b *testing.B) {
	createSolveBench2048A2(100000000, b)
	return
}

// func BenchmarkMemoryBreak2048A2(b *testing.B) {
// 	createSolveBench2048A2(50000000000, b)
// 	return
// }

func TestConcurrentMillion2048A2(t *testing.T) {
	createSolveConcurrent(1000000, t)
	return
}

func BenchmarkConcurrentManyMillions2048A2(b *testing.B) {
	createSolveConcurrentNBench(10000000, runtime.NumCPU(), b)
	return
}

// 4748 0975 4727 2012 8661 7503 4130 6167 7388 5051 2607 4492 0056 4448 6710
// 6196 3607 1042 4558 1476 5425 2707 6049 4101 2311 7758 9201 2567 5790 6462
// ...
// 0642 1926 9454 1125 0658 7397 7
// func Benchmark2019CheckPoint(b *testing.B) {
// 	puzzle := PuzzleRSW{
// 		N: new(big.Int).SetString("474809754727201286617503413061677388505126074492005644486710", 10),
// 	}

// 	return
// }
