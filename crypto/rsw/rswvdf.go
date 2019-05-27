package rsw

import "github.com/mit-dci/opencx/crypto"

// NewVDF is a randomized algorithm that takes a security parameter lambda and a desired puzzle difficulty t and produces a VDF construction based on RSW96, as well as public parameters.
// As far as the formal definition is concerned, this is the Setup(lambda, t) -> pp = (ek, vk) algorithm
func NewVDF(keySize uint64, t uint64) (rswvdf crypto.VDF, err error) {
	// TODO

	return
}

func (pz *PuzzleRSW) Eval(x []byte) (y []byte, proof []byte) {
	// TODO

	return
}

func (pz *PuzzleRSW) Verify(proof []byte, x []byte, y []byte) (valid bool) {
	// TODO

	return
}
