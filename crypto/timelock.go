package crypto

// Timelock is an interface that all timelock implementations should conform to.
type Timelock interface {
	// Setup sends key k to the future in time t, returning a puzzle and an answer
	SetupTimelock(t uint64) (*Puzzle, *Answer, error)
}

// Puzzle is what can actually be solved. It should return the same time t
type Puzzle interface {
	Solve() (*Answer, error)
}

// Answer is just a wrapper to make the interfaces make a little bit more sense.
// I'll remove this and replace with []byte if it's annoying
type Answer []byte
