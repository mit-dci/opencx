package rsw

import (
	"encoding/binary"
	"math/big"

	"github.com/mit-dci/opencx/crypto"
)

// NewVDF is a randomized algorithm that takes a security parameter lambda and a desired puzzle difficulty t and produces a VDF construction based on RSW96, as well as public parameters.
// As far as the formal definition is concerned, this is the Setup(lambda, t) -> pp = (ek, vk) algorithm
func NewVDF(keySize, k, t uint64) (rswvdf crypto.VDF, err error) {
	// TODO
	// we'll do the fiat-shamir transform version of this
	// l := H_prime(concatBytes(g.Bytes(), y.Bytes()))
	// H_prime: hash onto Primes(2k)
	// pi := g^(floor(2^t / l))
	panic("NewVDF(keySize, k, t uint64) is TODO")

	return
}

func (pz *PuzzleRSW) Eval(x []byte) (y []byte, proof []byte) {
	// TODO
	panic("PuzzleRSW Eval(x []byte) is TODO")

	return
}

func (pz *PuzzleRSW) Verify(proof, x, y []byte) (valid bool) {
	// TODO
	panic("PuzzleRSW Verify(proof, x, y []byte) is TODO")

	return
}

// The hash function in the paper is H_G(x) = int(H("residue"||x)) mod N (where H is a secure cryptographic hash function)

func HashOntoPrimes(twoPower uint64) (prime *big.Int, err error) {
	// TODO
	panic("HashOntoPrimes(twoPower uint64) is TODO")

	// Just put the number into a byte buffer
	finalBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(finalBuf, twoPower)

	// hasher := sha3.New256()

	return
}
