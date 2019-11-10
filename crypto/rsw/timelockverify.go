package rsw

import (
	"bytes"
	"fmt"
	"math/big"
)

// VerifyPuzzleOutput verifies that the timelock puzzle PuzzleRSW
func VerifyPuzzleOutput(p *big.Int, q *big.Int, pz *PuzzleRSW, claimedKey []byte) (valid bool, err error) {
	if p == nil {
		err = fmt.Errorf("p pointer cannot be nil, please investigate")
		return
	}
	if q == nil {
		err = fmt.Errorf("q pointer cannot be nil, please investigate")
		return
	}

	tempN := new(big.Int)
	tempN.Mul(p, q)
	if tempN.Cmp(pz.N) != 0 {
		err = fmt.Errorf("The p and q given do not multiply to the puzzle modulus")
		return
	}

	// compute trapdoor
	// phi(n) = (p-1)(q-1). We assume p and q are prime, and n = pq.
	phi := new(big.Int).Mul(new(big.Int).Sub(p, big.NewInt(1)), new(big.Int).Sub(q, big.NewInt(1)))

	// e = 2^t mod phi()
	e := new(big.Int).Exp(big.NewInt(2), pz.T, phi)

	// b = a^(e()) (mod n())
	b := new(big.Int).Exp(pz.A, e, pz.N)

	// now xor with ck, getting the bytes
	// if this is xor then the ck, err = blah like needs to be xor as well
	var answer []byte
	xorBytes := new(big.Int).Xor(pz.CK, b).Bytes()
	if len(xorBytes) <= 16 {
		answerBacking := [16]byte{}
		answer = answerBacking[:]
	} else {
		answer = make([]byte, len(xorBytes))
	}
	copy(answer, xorBytes)

	if res := bytes.Compare(answer, claimedKey); res != 0 {
		err = fmt.Errorf("The claimed key:\n\t%x\nIs not equal to the puzzle solution:\n\t%x\nSo the claimed solution is invalid", claimedKey, answer)
		return
	}

	valid = true
	return
}
