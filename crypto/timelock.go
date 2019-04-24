package crypto

// Timelock is an interface that all timelock implementations should conform to.
type Timelock interface {
	// SetupTimelock sends key k to the future in time t, returning a puzzle and an answer, or fails
	SetupTimelock(t uint64) (puzzle *Puzzle, answer []byte, err error)
}

// Puzzle is what can actually be solved. It should return the same time t
type Puzzle interface {
	// Solve solves the puzzle and returns the answer, or fails
	Solve() (answer []byte, err error)
}
