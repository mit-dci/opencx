package crypto

// VDF is the interface for verifiable delay functions
type VDF interface {
	// Eval takes an input x, and produces an output y and a proof
	Eval(x []byte) (y []byte, proof []byte)
	// Verify takes in a proof, x, and y, and will output the validity of the proof based on the VDF construction.
	Verify(proof []byte, x []byte, y []byte) (valid bool)
}
