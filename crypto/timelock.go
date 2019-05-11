package crypto

// Timelock is an interface that all timelock implementations should conform to.
type Timelock interface {
	// SetupTimelockPuzzle sends key k to the future in time t, returning a puzzle and an answer, or fails
	SetupTimelockPuzzle(t uint64) (puzzle Puzzle, answer []byte, err error)
}

// Puzzle is what can actually be solved. It should return the same answer that was the result of SetupTimelockPuzzle.
type Puzzle interface {
	// Solve solves the puzzle and returns the answer, or fails
	Solve() (answer []byte, err error)
	// Serialize turns the puzzle into something that's able to be sent over the wire
	Serialize() (raw []byte, err error)
}
