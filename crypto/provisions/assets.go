package provisions

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
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
	ti *big.Int
	xi *big.Int
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
func (machine *BalProofMachine) SetChallenge(ci *big.Int) (err error) {
	if ci == nil {
		err = fmt.Errorf("Challenge cannot be nil")
		return
	}
	machine.ci = ci
	return
}

// SetPrivKey sets the private key for this iteration
func (machine *BalProofMachine) SetPrivKey(privkey *ecdsa.PrivateKey) (err error) {
	if privkey == nil {
		err = fmt.Errorf("Private key cannot be nil")
		return
	}
	machine.xi = privkey.D
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
	// This is the anonymity set PK in the paper, !(privkey == nil) => s_i = 0
	// We don't use a privkey set because we are going to want to index by pubkey a lot
	PubKeyAnonSet    map[*ecdsa.PublicKey]*ecdsa.PrivateKey
	pkAnonSetMutex   *sync.Mutex
	BalanceRetreiver func(pubkey *ecdsa.PublicKey) (balance uint64, err error)
}

// NewAssetsProofMachine creates a new state machine for the asset proof. TODO: Use the anonymity set choosing from
// the paper to initialize
func NewAssetsProofMachine(curve elliptic.Curve) (machine *AssetsProofMachine, err error) {

	machine = &AssetsProofMachine{
		curve:            curve,
		PubKeyAnonSet:    make(map[*ecdsa.PublicKey]*ecdsa.PrivateKey),
		pkAnonSetMutex:   new(sync.Mutex),
		BalanceRetreiver: bal,
	}

	return
}

// bal is supposed to find the balance of a pubkey. I'm going to assume that all addresses have 10 bitcoin (1000000000)
// in them for now. TODO: Make bal() work in a production environment
func bal(pubkey *ecdsa.PublicKey) (bal uint64, err error) {

	// all pubkeys have 10 btc -- we measure in satoshis
	bal = 1000000000

	return
}

// calculateAssets calculates the assets that the exchange owns. This will be committed to.
func (machine *AssetsProofMachine) calculateAssets() (totalAssets uint64, err error) {

	machine.pkAnonSetMutex.Lock()
	for pub, priv := range machine.PubKeyAnonSet {
		// priv == nil => s_i == 0
		// so if priv != nil then add to the total assets
		if priv != nil {
			var currBal uint64
			if currBal, err = machine.BalanceRetreiver(pub); err != nil {
				machine.pkAnonSetMutex.Unlock()
				err = fmt.Errorf("Error getting balance of pubkey while calulating assets: %s", err)
				return
			}
			totalAssets += currBal
		}
		// otherwise don't add anything
	}

	machine.pkAnonSetMutex.Unlock()

	return
}
