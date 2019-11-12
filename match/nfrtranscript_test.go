package match

import (
	"fmt"
	"testing"

	"github.com/mit-dci/lit/crypto/koblitz"
	"golang.org/x/crypto/sha3"
)

// TestEmptyTranscripVerify tests an empty transcript and makes sure
// that it is not valid
func TestEmptyTranscripVerify(t *testing.T) {
	var err error
	emptyTranscript := Transcript{}

	var valid bool
	valid, err = emptyTranscript.Verify()

	if valid {
		t.Errorf("Empty transcript should have been invalid, was valid")
		return
	}
	return
}

// TestOneOrderValidTranscripVerify creates a transcript with a single
// order in it and tests that it is valid.
func TestOneOrderValidTranscripVerify(t *testing.T) {
	var err error

	// create exchange private key
	var exprivkey *koblitz.PrivateKey
	if exprivkey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		t.Errorf("Error creating exchange private key for signing: %s", err)
		return
	}

	// init empty transcript, the id from there is valid
	emptyTranscript := Transcript{}

	var idsig *koblitz.Signature
	if idsig, err = exprivkey.Sign(hash256(emptyTranscript.batchId[:])); err != nil {
		err = fmt.Errorf("Error with exchange signing batch id: %s", err)
		return
	}
	emptyTranscript.batchIdSig = idsig.Serialize()

	// set time param
	time := 100000

	// NOTE: start of user stuff
	// First create solution
	var soln SolutionOrder
	if soln, err = NewSolutionOrder(2048); err != nil {
		t.Errorf("Error creating solution order of 2048 bits: %s", err)
		return
	}

	// now create encrypted order NOTE: change t to be massive on
	// larger tests
	var encOrder EncryptedSolutionOrder
	if encOrder, err = soln.EncryptSolutionOrder(origOrder, time); err != nil {
		t.Errorf("Error encrypting solution order for test: %s", err)
		return
	}

	var valid bool
	valid, err = emptyTranscript.Verify()

	if valid {
		t.Errorf("Empty transcript should have been invalid, was valid")
		return
	}
	return
}

// hash256 takes sha3 256-bit hash of some bytes - this ignores
// errors.
func hash256(preimage []byte) (h []byte) {
	hashingAlgo := sha3.New256()
	hashingAlgo.Write(preimage)
	h = hashingAlgo.Sum(nil)
	return
}
