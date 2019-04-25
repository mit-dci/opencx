package rsw

import (
	"fmt"
	"math/big"

	"github.com/mit-dci/opencx/crypto"
)

// TimelockRSW generates the puzzle that can then only be solved with repeated squarings
type TimelockRSW struct {
	p *big.Int
	q *big.Int
	t *big.Int
	a *big.Int
}

// PuzzleRSW is the puzzle that can be then solved by repeated modular squaring
type PuzzleRSW struct {
	n  *big.Int
	a  *big.Int
	t  *big.Int
	ck *big.Int
	cm *big.Int
}

func (tl *TimelockRSW) n() (n *big.Int, err error) {
	if tl.p == nil || tl.q == nil {
		err = fmt.Errorf("Must set up p and q to get n")
		return
	}
	// n = pq
	n.Mul(tl.p, tl.q)
	return
}

// ϕ() = phi(n) = (p-1)(q-1)
func (tl *TimelockRSW) ϕ() (ϕ *big.Int, err error) {
	if tl.p == nil || tl.q == nil {
		err = fmt.Errorf("Must set up p and q to get the ϕ")
		return
	}
	// ϕ(n) = (p-1)(q-1). We assume p and q are prime, and n = pq.
	ϕ.Mul(tl.p.Sub(tl.p, big.NewInt(int64(1))), tl.q.Sub(tl.q, big.NewInt(1)))
	return
}

// e = 2^t (mod ϕ()) = 2^t (mod phi(n))
func (tl *TimelockRSW) e() (e *big.Int, err error) {
	if tl.t == nil {
		err = fmt.Errorf("Must set up t in order to get e")
		return
	}
	var ϕ *big.Int
	if ϕ, err = tl.ϕ(); err != nil {
		err = fmt.Errorf("Could not find ϕ: %s", err)
		return
	}
	// e = 2^t mod ϕ()
	e.Exp(big.NewInt(int64(2)), tl.t, ϕ)
	return
}

// b = a^(e()) (mod n()) = a^e (mod n) = a^(2^t (mod ϕ())) (mod n) = a^(2^t) (mod n)
func (tl *TimelockRSW) b() (b *big.Int, err error) {
	if tl.a == nil {
		err = fmt.Errorf("Must set up a and n in order to get b")
		return
	}
	var n *big.Int
	if n, err = tl.n(); err != nil {
		err = fmt.Errorf("Could not find n: %s", err)
		return
	}
	var e *big.Int
	if e, err = tl.e(); err != nil {
		err = fmt.Errorf("Could not find e: %s", err)
		return
	}
	// b = a^(e()) (mod n())
	b.Exp(tl.a, e, n)
	return
}

// SetupTimelockPuzzle sets up the time lock puzzle for the scheme described in RSW96
func (tl *TimelockRSW) SetupTimelockPuzzle(t uint64) (puzzle crypto.Puzzle, err error) {
	return
}
