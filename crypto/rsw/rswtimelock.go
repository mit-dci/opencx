// Package rsw is an implementation of Rivest-Shamir-Wagner timelock puzzles, from RSW96.
// The puzzles are based on modular exponentiation, and this package provides an easy to use API for creating and solving these types of puzzles.
package rsw

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"
	"fmt"

	"math/big"

	// danbig "github.com/Rjected/gmp"
	gmpbig "github.com/Rjected/gmp"

	"github.com/mit-dci/opencx/crypto"
)

// TimelockRSW generates the puzzle that can then only be solved with repeated squarings
type TimelockRSW struct {
	rsaKeyBits int
	key        []byte
	p          *big.Int
	q          *big.Int
	t          *big.Int
	a          *big.Int
}

// PuzzleRSW is the puzzle that can be then solved by repeated modular squaring
type PuzzleRSW struct {
	N *big.Int
	A *big.Int
	T *big.Int
	// We use C_k = b xor k
	CK *big.Int
}

// New creates a new TimelockRSW with p and q generated as per crypto/rsa, and an input a as well as number of bits for the RSA key size.
// The key is also set here
// The number of bits is so we can figure out how big we want p and q to be.
func New(key []byte, a int64, rsaKeyBits int) (timelock crypto.Timelock, err error) {
	tl := new(TimelockRSW)
	tl.rsaKeyBits = rsaKeyBits
	// generate primes p and q
	var rsaPrivKey *rsa.PrivateKey
	if rsaPrivKey, err = rsa.GenerateMultiPrimeKey(rand.Reader, 2, tl.rsaKeyBits); err != nil {
		err = fmt.Errorf("Could not generate primes for RSA: %s", err)
		return
	}
	if len(rsaPrivKey.Primes) != 2 {
		err = fmt.Errorf("For some reason the RSA Privkey has != 2 primes, this should not be the case for RSW, we only need p and q")
		return
	}
	// If we ever want to just switch to gmp for all calculations these two lines will fix all of the issues
	tl.p = new(big.Int).SetBytes(rsaPrivKey.Primes[0].Bytes())
	tl.q = new(big.Int).SetBytes(rsaPrivKey.Primes[1].Bytes())
	tl.a = big.NewInt(a)
	tl.key = key

	timelock = tl
	return
}

// New2048 creates a new TimelockRSW with p and q generated as per crypto/rsa, and an input a. This generates according to a fixed RSA key size (2048 bits).
func New2048(key []byte, a int64) (tl crypto.Timelock, err error) {
	return New(key, a, 2048)
}

// New2048A2 is the same as New2048 but we use a base of 2. It's called A2 because A=2 I guess
func New2048A2(key []byte) (tl crypto.Timelock, err error) {
	return New(key, 2, 2048)
}

// NewTimelockWithPrimes creates a new timelock puzzle with primes p
// and q.
func NewTimelockWithPrimes(key []byte, a uint64, p *big.Int, q *big.Int) (timelock crypto.Timelock, err error) {
	if p == nil {
		err = fmt.Errorf("p pointer cannot be nil, please investigate")
		return
	}
	if q == nil {
		err = fmt.Errorf("q pointer cannot be nil, please investigate")
		return
	}
	tl := new(TimelockRSW)
	aInt := int64(a)

	// Create a temporary N = p * q and set the bit length
	tempN := new(big.Int)
	tempN.Mul(p, q)
	tl.rsaKeyBits = tempN.BitLen()

	// now set the p and q values
	tl.p = new(big.Int)
	tl.q = new(big.Int)
	tl.p.Set(p)
	tl.q.Set(q)

	// Set the a value
	tl.a = big.NewInt(aInt)

	// copy over the key
	tl.key = make([]byte, len(key))
	copy(tl.key, key)

	timelock = tl
	return
}

func (tl *TimelockRSW) n() (n *big.Int, err error) {
	if tl.p == nil || tl.q == nil {
		err = fmt.Errorf("Must set up p and q to get n")
		return
	}
	// n = pq
	n = new(big.Int).Mul(tl.p, tl.q)
	return
}

// phi() = phi(n) = (p-1)(q-1)
func (tl *TimelockRSW) phi() (phi *big.Int, err error) {
	if tl.p == nil || tl.q == nil {
		err = fmt.Errorf("Must set up p and q to get the phi")
		return
	}
	// phi(n) = (p-1)(q-1). We assume p and q are prime, and n = pq.
	phi = new(big.Int).Mul(new(big.Int).Sub(tl.p, big.NewInt(1)), new(big.Int).Sub(tl.q, big.NewInt(1)))
	return
}

// e = 2^t (mod phi()) = 2^t (mod phi(n))
func (tl *TimelockRSW) e() (e *big.Int, err error) {
	if tl.t == nil {
		err = fmt.Errorf("Must set up t in order to get e")
		return
	}
	var phi *big.Int
	if phi, err = tl.phi(); err != nil {
		err = fmt.Errorf("Could not find phi: %s", err)
		return
	}
	// e = 2^t mod phi()
	e = new(big.Int).Exp(big.NewInt(2), tl.t, phi)
	return
}

// b = a^(e()) (mod n()) = a^e (mod n) = a^(2^t (mod phi())) (mod n) = a^(2^t) (mod n)
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
	b = new(big.Int).Exp(tl.a, e, n)
	return
}

func (tl *TimelockRSW) ckXOR() (ck *big.Int, err error) {
	var b *big.Int
	if b, err = tl.b(); err != nil {
		err = fmt.Errorf("Could not find b: %s", err)
		return
	}

	// set k to be the bytes of the key
	k := new(big.Int).SetBytes(tl.key)

	// C_k = k ⊕ a^(2^t) (mod n) = k ⊕ b (mod n)
	ck = new(big.Int).Xor(b, k)
	return
}

func (tl *TimelockRSW) ckADD() (ck *big.Int, err error) {
	var b *big.Int
	if b, err = tl.b(); err != nil {
		err = fmt.Errorf("Could not find b: %s", err)
		return
	}
	// set k to be the bytes of the key
	k := new(big.Int).SetBytes(tl.key)

	// C_k = k + a^(2^t) (mod n) = k + b (mod n)
	// TODO: does this need to be ck.Add(b, k).Mod(ck, n)?
	ck = new(big.Int).Add(b, k)
	return
}

// SetupTimelockPuzzle sets up the time lock puzzle for the scheme described in RSW96. This uses the normal crypto/rsa way of selecting primes p and q.
// You should throw away the answer but some puzzles like the hash puzzle make sense to have the answer as an output of the setup, since that's the decryption key and you don't know beforehand how to encrypt.
func (tl *TimelockRSW) SetupTimelockPuzzle(t uint64) (puzzle crypto.Puzzle, answer []byte, err error) {
	tl.t = new(big.Int).SetUint64(t)
	var n *big.Int
	if n, err = tl.n(); err != nil {
		err = fmt.Errorf("Could not find n: %s", err)
		return
	}
	// if this is xor then the answer = blah line needs to be xor as well
	var ck *big.Int
	if ck, err = tl.ckXOR(); err != nil {
		err = fmt.Errorf("Could not find ck: %s", err)
		return
	}

	rswPuzzle := &PuzzleRSW{
		N:  n,
		A:  tl.a,
		T:  tl.t,
		CK: ck,
	}
	puzzle = rswPuzzle

	var b *big.Int
	if b, err = tl.b(); err != nil {
		err = fmt.Errorf("Could not find b: %s", err)
		return
	}

	// if this is xor then the ck, err = blah like needs to be xor as well
	xorBytes := new(big.Int).Xor(ck, b).Bytes()
	if len(xorBytes) <= 16 {
		answerBacking := [16]byte{}
		answer = answerBacking[:]
	} else {
		answer = make([]byte, len(xorBytes))
	}
	copy(answer, xorBytes)
	return
}

// SolveCkADD solves the puzzle by repeated squarings and subtracting b from ck
func (pz *PuzzleRSW) SolveCkADD() (answer []byte, err error) {
	// Make sure that the answer is 16 bytes long, padded
	ansBytes := new(big.Int).Sub(pz.CK, new(big.Int).Exp(pz.A, new(big.Int).Exp(big.NewInt(2), pz.T, nil), pz.N)).Bytes()
	if len(ansBytes) <= 16 {
		answer = make([]byte, 16)
	} else {
		answer = make([]byte, len(ansBytes))
	}
	copy(answer, ansBytes)
	return
}

// SolveCkXOR solves the puzzle by repeated squarings and xor b with ck
func (pz *PuzzleRSW) SolveCkXOR() (answer []byte, err error) {
	// Make sure that the answer is 16 bytes long, padded
	ansBytes := new(big.Int).Xor(pz.CK, new(big.Int).Exp(pz.A, new(big.Int).Exp(big.NewInt(2), pz.T, nil), pz.N)).Bytes()
	if len(ansBytes) <= 16 {
		answer = make([]byte, 16)
	} else {
		answer = make([]byte, len(ansBytes))
	}
	copy(answer, ansBytes)
	return
}

// Solve solves the puzzle by repeated squarings
func (pz *PuzzleRSW) Solve() (answer []byte, err error) {
	// return pz.SolveDanGMPCkXOR()
	return pz.SolveGMPCkXOR()
}

// SolveGMPCkXOR solves the puzzle by repeated squarings and xor b with ck using the GMP library
func (pz *PuzzleRSW) SolveGMPCkXOR() (answer []byte, err error) {
	// No longer a one liner but many times faster
	// return new(gmpbig.Int).Xor(new(gmpbig.Int).SetBytes(pz.CK.Bytes()), new(gmpbig.Int).Exp(new(gmpbig.Int).SetBytes(pz.A.Bytes()), new(gmpbig.Int).Exp(gmpbig.NewInt(2), new(gmpbig.Int).SetBytes(pz.T.Bytes()), nil), new(gmpbig.Int).SetBytes(pz.N.Bytes()))).Bytes(), nil
	// we're using a fork now!
	// make sure it's 16 bytes long padded
	ansBytes :=
		new(gmpbig.Int).Xor(new(gmpbig.Int).SetBytes(pz.CK.Bytes()),
			new(gmpbig.Int).ExpSquare(new(gmpbig.Int).SetBytes(pz.A.Bytes()),
				new(gmpbig.Int).SetBytes(pz.T.Bytes()),
				new(gmpbig.Int).SetBytes(pz.N.Bytes()))).Bytes()
	if len(ansBytes) <= 16 {
		answer = make([]byte, 16)
	} else {
		answer = make([]byte, len(ansBytes))
	}
	copy(answer, ansBytes)
	return
}

// SolveGMPCkADD solves the puzzle by repeated squarings and xor b with ck using the GMP library
func (pz *PuzzleRSW) SolveGMPCkADD() (answer []byte, err error) {
	// No longer a one liner but many times faster
	gmpck := new(gmpbig.Int).SetBytes(pz.CK.Bytes())
	gmpa := new(gmpbig.Int).SetBytes(pz.A.Bytes())
	gmpt := new(gmpbig.Int).SetBytes(pz.T.Bytes())
	gmpn := new(gmpbig.Int).SetBytes(pz.N.Bytes())
	// The answer is padded when we create it, so it should be padded when we solve
	ansBytes := new(gmpbig.Int).Sub(gmpck, new(gmpbig.Int).ExpSquare(gmpa, gmpt, gmpn)).Bytes()
	if len(ansBytes) <= 16 {
		answer = make([]byte, 16)
	} else {
		answer = make([]byte, len(ansBytes))
	}
	copy(answer, ansBytes)
	return
}

// Serialize turns the RSW puzzle into something that can be sent over the wire
func (pz *PuzzleRSW) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register puzzleRSW interface
	gob.Register(PuzzleRSW{})

	// create a new encoder writing to the buffer
	enc := gob.NewEncoder(&b)

	// encode the puzzle in the buffer
	if err = enc.Encode(pz); err != nil {
		err = fmt.Errorf("Error encoding puzzle: %s", err)
		return
	}

	// Get the bytes from the buffer
	raw = b.Bytes()

	return
}

// Deserialize turns a gob-encoded puzzle into a go struct we can use.
func (pz *PuzzleRSW) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register puzzleRSW interface
	gob.Register(PuzzleRSW{})

	// create a new decoder writing to the buffer
	dec := gob.NewDecoder(b)

	// decode the puzzle in the buffer
	if err = dec.Decode(pz); err != nil {
		err = fmt.Errorf("Error decoding puzzle: %s", err)
		return
	}

	return
}
