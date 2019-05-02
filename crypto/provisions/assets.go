package provisions

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

// BalProofMachine represents a single iteration of the proof of assets protocol, and is
// used to compute individual balance commitments, as well as calculate things like
// responses to challenges
type BalProofMachine struct {
	u1 *big.Int
	u2 *big.Int
	u3 *big.Int
	u4 *big.Int
	ci *big.Int
}

// NewBalProofMachine creates a new balance proof machine
func NewBalProofMachine(curve elliptic.Curve) (machine *BalProofMachine, err error) {
	order := curve.Params().P
	if machine.u1, err = rand.Int(rand.Reader, order); err != nil {
		err = fmt.Errorf("Error getting random u_1 for balance proof machine: %s", err)
		return
	}

	if machine.u2, err = rand.Int(rand.Reader, order); err != nil {
		err = fmt.Errorf("Error getting random u_2 for balance proof machine: %s", err)
		return
	}

	if machine.u3, err = rand.Int(rand.Reader, order); err != nil {
		err = fmt.Errorf("Error getting random u_3 for balance proof machine: %s", err)
		return
	}

	if machine.u4, err = rand.Int(rand.Reader, order); err != nil {
		err = fmt.Errorf("Error getting random u_4 for balance proof machine: %s", err)
		return
	}

	return
}

// SetChallenge sets the challenge so we can generate responses
func (machine *BalProofMachine) SetChallenge(ci *big.Int) {
	machine.ci = ci
	return
}

// SResponse generates the response r_(s_i) with the balance proof machine and si (s_i). The challenge must be set.
func (machine *BalProofMachine) SResponse(si bool) (rs *big.Int, err error) {
	if machine.ci == nil {
		err = fmt.Errorf("Cannot generate a reponse to a challenge if the challenge has not been set")
		return
	}

	if si {
		rs = new(big.Int).Add(machine.u1, machine.ci)
	} else {
		rs = new(big.Int).Set(machine.u1)
	}

	return
}

// AssetsProofMachine is the state machine that is used to create a privacy preserving proof of assets
type AssetsProofMachine struct {
	curve elliptic.Curve
	// PrivKeySet is the set of all x_i
	PrivKeySet []*ecdsa.PrivateKey
}
