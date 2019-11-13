package match

import (
	"fmt"
	"testing"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
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

func runBenchTranscriptVerify(b *testing.B, time uint64, orders uint64) {
	var err error
	if orders == 0 {
		b.Fatalf("Cannot run test with no orders, please setup test correctly")
		return
	}

	b.StopTimer()
	logging.SetLogLevel(3)
	// create exchange private key
	var exprivkey *koblitz.PrivateKey
	if exprivkey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		b.Fatalf("Error creating exchange private key for signing: %s", err)
		return
	}

	// init empty transcript, the id from there is valid
	transcript := Transcript{}

	// var idsig *koblitz.Signature
	// if idsig, err = exprivkey.Sign(hash256(transcript.batchId[:])); err != nil {
	// 	err = fmb.Fatalf("Error with exchange signing batch id: %s", err)
	// 	return
	// }
	// transcript.batchIdSig = idsig.Serialize()
	hasher := sha3.New256()
	hasher.Write(transcript.batchId[:])
	var batchSig []byte
	if batchSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, hasher.Sum(nil), false); err != nil {
		b.Fatalf("Error compact signing batch id sig: %s", err)
		return
	}
	transcript.batchIdSig = make([]byte, len(batchSig))
	copy(transcript.batchIdSig, batchSig)

	hasher.Reset()
	// This maps private key to solution order so we can respond
	// correctly later.
	var privkeyOrderMap map[koblitz.PrivateKey]SolutionOrder = make(map[koblitz.PrivateKey]SolutionOrder)
	for i := uint64(0); i < orders; i++ {
		// NOTE: start of user stuff
		var userPrivKey *koblitz.PrivateKey
		if userPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			b.Fatalf("Error creating new user private key for signing: %s", err)
			return
		}

		// First create solution
		var soln SolutionOrder
		if soln, err = NewSolutionOrder(2048); err != nil {
			b.Fatalf("Error creating solution order of 2048 bits: %s", err)
			return
		}
		privkeyOrderMap[*userPrivKey] = soln

		// now create encrypted order NOTE: change t to be massive on
		// larger tests
		var encOrder EncryptedSolutionOrder
		if encOrder, err = soln.EncryptSolutionOrder(*origOrder, time); err != nil {
			b.Fatalf("Error encrypting solution order for test: %s", err)
			return
		}

		var encOrderBuf []byte
		if encOrderBuf, err = encOrder.Serialize(); err != nil {
			b.Fatalf("Error serializing encrypted order before signing: %s", err)
			return
		}

		hasher.Reset()
		hasher.Write(encOrderBuf)
		var userSigBuf []byte
		if userSigBuf, err = koblitz.SignCompact(koblitz.S256(), userPrivKey, hasher.Sum(nil), false); err != nil {
			b.Fatalf("Error signing encrypted order for user: %s", err)
			return
		}

		// now that we've created the solution order we add it to the
		// transcript as being "submitted".
		signedOrder := SignedEncSolOrder{
			EncSolOrder: encOrder,
			Signature:   make([]byte, len(userSigBuf)),
		}

		// NOTE: this is the most likely point of failure
		copy(signedOrder.Signature, userSigBuf)
		transcript.puzzledOrders = append(transcript.puzzledOrders, signedOrder)
	}

	// now that we have a bunch of puzzled orders, we should create a
	// commitment out of it.
	hasher.Reset()
	for _, encOrder := range transcript.puzzledOrders {
		var rawPuzzle []byte
		if rawPuzzle, err = encOrder.Serialize(); err != nil {
			b.Fatalf("Error serializing submitted order before hashing: %s", err)
			return
		}
		hasher.Write(rawPuzzle)
	}
	copy(transcript.commitment[:], hasher.Sum(nil))
	var exchangeCommSig []byte
	if exchangeCommSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, transcript.commitment[:], false); err != nil {
		b.Fatalf("Error with exchange signing the commitment: %s", err)
		return
	}
	transcript.commitSig = make([]byte, len(exchangeCommSig))
	copy(transcript.commitSig, exchangeCommSig)

	hasher.Reset()
	// users now create their signatures and reveal solutions
	for userprivkey, solnorder := range privkeyOrderMap {
		// because we're running a test we do not check the time -- in
		// reality you NEED to check the time elapsed.
		userCommitResponse := CommitResponse{}

		// h(commit + sig + answer) = e
		hasher.Reset()
		hasher.Write(transcript.commitment[:])
		hasher.Write(transcript.commitSig)
		var solnOrderBuf []byte
		if solnOrderBuf, err = solnorder.Serialize(); err != nil {
			b.Fatalf("Error serializing solution order for response: %s", err)
			return
		}
		hasher.Write(solnOrderBuf)
		var responseSigBuf []byte
		if responseSigBuf, err = koblitz.SignCompact(koblitz.S256(), &userprivkey, hasher.Sum(nil), false); err != nil {
			b.Fatalf("Error for user signing response: %s", err)
			return
		}
		if len(responseSigBuf) != 65 {
			b.Fatalf("Error in test: response signature is not 65 bytes")
			return
		}
		copy(userCommitResponse.CommResponseSig[:], responseSigBuf)
		userCommitResponse.PuzzleAnswerReveal = solnorder
	}

	b.StartTimer()
	b.ResetTimer()
	var valid bool
	valid, err = transcript.Verify()

	if !valid {
		logging.Fatalf("Error from benchmark: %s", err)
		b.Fatalf("Empty transcript should have been valid, was invalid: %s", err)
		return
	}
	return
}

func BenchmarkValidTranscript(b *testing.B) {
	orderAmounts := []uint64{1, 10, 100}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("BenchValidNFR%d", amt), func(g *testing.B) {
			for i := 0; i < g.N; i++ {
				runBenchTranscriptVerify(b, 10000, amt)
			}
		})
	}
}

// func BenchmarkValidTranscript10_0(b *testing.B) {
// 	runBenchTranscriptVerify(b, 10000, 1)
// }

// func BenchmarkValidTranscript10_1(b *testing.B) {
// 	runBenchTranscriptVerify(b, 10000, 10)
// }

// func BenchmarkValidTranscript10_2(b *testing.B) {
// 	runBenchTranscriptVerify(b, 10000, 100)
// }

// func BenchmarkValidTranscript10_3(b *testing.B) {
// 	runBenchTranscriptVerify(b, 10000, 1000)
// }

func runValidTranscriptVerify(t *testing.T, time uint64, orders uint64) {
	var err error
	if orders == 0 {
		t.Errorf("Cannot run test with no orders, please setup test correctly")
		return
	}

	logging.SetLogLevel(3)
	// create exchange private key
	var exprivkey *koblitz.PrivateKey
	if exprivkey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		t.Errorf("Error creating exchange private key for signing: %s", err)
		return
	}

	// init empty transcript, the id from there is valid
	emptyTranscript := Transcript{}

	// var idsig *koblitz.Signature
	// if idsig, err = exprivkey.Sign(hash256(emptyTranscript.batchId[:])); err != nil {
	// 	err = fmt.Errorf("Error with exchange signing batch id: %s", err)
	// 	return
	// }
	// emptyTranscript.batchIdSig = idsig.Serialize()
	var batchSig []byte
	if batchSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, hash256(emptyTranscript.batchId[:]), false); err != nil {
		t.Errorf("Error compact signing batch id sig: %s", err)
		return
	}
	emptyTranscript.batchIdSig = make([]byte, len(batchSig))
	copy(emptyTranscript.batchIdSig, batchSig)

	// This maps private key to solution order so we can respond
	// correctly later.
	var privkeyOrderMap map[koblitz.PrivateKey]SolutionOrder = make(map[koblitz.PrivateKey]SolutionOrder)
	for i := uint64(0); i < orders; i++ {
		// NOTE: start of user stuff
		var userPrivKey *koblitz.PrivateKey
		if userPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			t.Errorf("Error creating new user private key for signing: %s", err)
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
		if encOrder, err = soln.EncryptSolutionOrder(*origOrder, time); err != nil {
			t.Errorf("Error encrypting solution order for test: %s", err)
			return
		}

		var encOrderBuf []byte
		if encOrderBuf, err = encOrder.Serialize(); err != nil {
			t.Errorf("Error serializing encrypted order before signing: %s", err)
			return
		}

		var userSigBuf []byte
		if userSigBuf, err = koblitz.SignCompact(koblitz.S256(), userPrivKey, hash256(encOrderBuf), false); err != nil {
			t.Errorf("Error signing encrypted order for user: %s", err)
			return
		}

		// now that we've created the solution order we add it to the
		// transcript as being "submitted".
		signedOrder := SignedEncSolOrder{
			EncSolOrder: encOrder,
			Signature:   make([]byte, len(userSigBuf)),
		}

		// NOTE: this is the most likely point of failure
		copy(signedOrder.Signature, userSigBuf)
		emptyTranscript.puzzledOrders = append(emptyTranscript.puzzledOrders, signedOrder)
	}

	// now that we have a bunch of puzzled orders, we should create a
	// commitment out of it.
	var commitmentPreImg []byte
	for _, encOrder := range emptyTranscript.puzzledOrders {
		var rawPuzzle []byte
		if rawPuzzle, err = encOrder.Serialize(); err != nil {
			t.Errorf("Error serializing submitted order before hashing: %s", err)
			return
		}
		commitmentPreImg = append(commitmentPreImg, rawPuzzle...)
	}
	copy(emptyTranscript.commitment[:], hash256(commitmentPreImg))
	var exchangeCommSig []byte
	if exchangeCommSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, emptyTranscript.commitment[:], false); err != nil {
		t.Errorf("Error with exchange signing the commitment: %s", err)
		return
	}
	emptyTranscript.commitSig = make([]byte, len(exchangeCommSig))
	copy(emptyTranscript.commitSig, exchangeCommSig)

	// users now create their signatures and reveal solutions
	for userprivkey, solnorder := range privkeyOrderMap {
		// because we're running a test we do not check the time -- in
		// reality you NEED to check the time elapsed.
		userCommitResponse := CommitResponse{}

		// h(commit + sig + answer) = e
		var responseBuf []byte
		responseBuf = append(responseBuf, emptyTranscript.commitment[:]...)
		responseBuf = append(responseBuf, emptyTranscript.commitSig...)
		var solnOrderBuf []byte
		if solnOrderBuf, err = solnorder.Serialize(); err != nil {
			t.Errorf("Error serializing solution order for response: %s", err)
			return
		}
		responseBuf = append(responseBuf, solnOrderBuf...)
		var responseSigBuf []byte
		if responseSigBuf, err = koblitz.SignCompact(koblitz.S256(), &userprivkey, hash256(responseBuf), false); err != nil {
			t.Errorf("Error for user signing response: %s", err)
			return
		}
		if len(responseSigBuf) != 65 {
			t.Errorf("Error in test: response signature is not 65 bytes")
			return
		}
		copy(userCommitResponse.CommResponseSig[:], responseSigBuf)
		userCommitResponse.PuzzleAnswerReveal = solnorder
	}

	var valid bool
	valid, err = emptyTranscript.Verify()

	if !valid {
		t.Errorf("Empty transcript should have been valid, was invalid: %s", err)
		return
	}
	return
}

// TestOneOrderValidTranscriptVerify creates a transcript with a single
// order in it and tests that it is valid.
func TestOneOrderValidTranscriptVerify(t *testing.T) {
	runValidTranscriptVerify(t, 10000, 1)
}

// hash256 takes sha3 256-bit hash of some bytes - this ignores
// errors.
func hash256(preimage []byte) (h []byte) {
	hashingAlgo := sha3.New256()
	hashingAlgo.Write(preimage)
	h = hashingAlgo.Sum(nil)
	return
}
