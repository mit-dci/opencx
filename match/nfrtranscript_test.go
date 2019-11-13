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
	emptyTranscript := Transcript{}

	// this should error so valid being true would also mean err ==
	// nil
	var valid bool
	valid, _ = emptyTranscript.Verify()

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

	// This maps private key to solution order so we can respond
	// correctly later.
	var privkeyOrderMap map[koblitz.PrivateKey]SolutionOrder = make(map[koblitz.PrivateKey]SolutionOrder)
	for i := 0; i < 100; i++ {
		// NOTE: start of user stuff
		var userPrivKey *koblitz.PrivateKey
		if userPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			err = fmt.Errorf("Error creating new user private key for signing: %s", err)
			return
		}

		// First create solution
		var soln SolutionOrder
		if soln, err = NewSolutionOrder(2048); err != nil {
			t.Errorf("Error creating solution order of 2048 bits: %s", err)
			return
		}
		privkeyOrderMap[*userPrivKey] = soln

		// now create encrypted order NOTE: change t to be massive on
		// larger tests
		var encOrder EncryptedSolutionOrder
		if encOrder, err = soln.EncryptSolutionOrder(*origOrder, uint64(time)); err != nil {
			t.Errorf("Error encrypting solution order for test: %s", err)
			return
		}

		var encOrderBuf []byte
		if encOrderBuf, err = encOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing encrypted order before signing: %s", err)
			return
		}

		// now we have to sign the order
		var solnOrderSig *koblitz.Signature
		if solnOrderSig, err = userPrivKey.Sign(hash256(encOrderBuf)); err != nil {
			err = fmt.Errorf("Error signing encrypted order for user: %s", err)
			return
		}

		// now that we've created the solution order we add it to the
		// transcript as being "submitted".
		signedOrder := SignedEncSolOrder{
			EncSolOrder: encOrder,
		}

		// NOTE: this is the most likely point of failure
		copy(signedOrder.Signature[:], solnOrderSig.Serialize())
		emptyTranscript.puzzledOrders = append(emptyTranscript.puzzledOrders, signedOrder)
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
