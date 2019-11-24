package match

import (
	"fmt"
	"math/big"
	"sync"
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

// runBenchTranscriptVerify runs a benchmark which creates orders with
// a time parameter specified by the user, and creates a valid
// transcript.
func runBenchTranscriptVerify(b *testing.B, time uint64, orders uint64) {
	// TODO: encrypt orders in parallel
	var err error
	if orders == 0 {
		b.Fatalf("Cannot run test with no orders, please setup test correctly")
		return
	}

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
	// if idsig, err = exprivkey.Sign(hash256(transcript.BatchId[:])); err != nil {
	// 	err = fmb.Fatalf("Error with exchange signing batch id: %s", err)
	// 	return
	// }
	// transcript.BatchIdSig = idsig.Serialize()
	hasher := sha3.New256()
	hasher.Write(transcript.BatchId[:])
	var batchSig []byte
	if batchSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, hasher.Sum(nil), false); err != nil {
		b.Fatalf("Error compact signing batch id sig: %s", err)
		return
	}
	transcript.BatchIdSig = make([]byte, len(batchSig))
	copy(transcript.BatchIdSig, batchSig)

	hasher.Reset()
	// This maps private key to solution order so we can respond
	// correctly later.
	var privkeyOrderMap map[koblitz.PrivateKey]SolutionOrder = make(map[koblitz.PrivateKey]SolutionOrder)
	// var bufMtx sync.Mutex
	var solnBuf []SolutionOrder = make([]SolutionOrder, orders)
	var errBuf []error = make([]error, orders)
	var wg sync.WaitGroup
	wg.Add(int(orders))
	for i := uint64(0); i < orders; i++ {
		// First create solution
		solnBuf[i].P = new(big.Int)
		solnBuf[i].Q = new(big.Int)
		go func(j int) {
			var soln SolutionOrder
			if soln, errBuf[j] = NewSolutionOrder(1024); errBuf[j] != nil {
				errBuf[j] = fmt.Errorf("Error creating solution order of 1024 bits: %s", errBuf[j])
				wg.Done()
				return
			}
			solnBuf[j].P.Set(soln.P)
			solnBuf[j].Q.Set(soln.Q)
			wg.Done()
		}(int(i))
	}
	wg.Wait()

	didError := false
	for j, creationErr := range errBuf {
		if creationErr != nil {
			didError = true
			b.Errorf("Error: the %d'th goroutine failed with error: %s", j, creationErr)
		}
	}
	if didError {
		b.FailNow()
	}

	for _, soln := range solnBuf {
		// NOTE: start of user stuff
		var userPrivKey *koblitz.PrivateKey
		if userPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			b.Fatalf("Error creating new user private key for signing: %s", err)
			return
		}

		privkeyOrderMap[*userPrivKey] = soln

		// now create encrypted order
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

		copy(signedOrder.Signature, userSigBuf)
		transcript.PuzzledOrders = append(transcript.PuzzledOrders, signedOrder)
	}

	// now that we have a bunch of puzzled orders, we should create a
	// commitment out of it.
	hasher.Reset()
	for _, encOrder := range transcript.PuzzledOrders {
		var rawPuzzle []byte
		if rawPuzzle, err = encOrder.Serialize(); err != nil {
			b.Fatalf("Error serializing submitted order before hashing: %s", err)
			return
		}
		hasher.Write(rawPuzzle)
	}
	copy(transcript.Commitment[:], hasher.Sum(nil))
	var exchangeCommSig []byte
	if exchangeCommSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, transcript.Commitment[:], false); err != nil {
		b.Fatalf("Error with exchange signing the commitment: %s", err)
		return
	}
	transcript.CommitSig = make([]byte, len(exchangeCommSig))
	copy(transcript.CommitSig, exchangeCommSig)

	hasher.Reset()

	// users now create their signatures and reveal solutions
	for userprivkey, solnorder := range privkeyOrderMap {
		// because we're running a test we do not check the time -- in
		// reality you NEED to check the time elapsed.
		userCommitResponse := CommitResponse{}

		// h(commit + sig + answer) = e
		hasher.Reset()
		hasher.Write(transcript.Commitment[:])
		hasher.Write(transcript.CommitSig)
		var solnOrderBuf []byte
		if solnOrderBuf, err = solnorder.Serialize(); err != nil {
			b.Fatalf("Error serializing solution order for response: %s", err)
			return
		}
		hasher.Write(solnOrderBuf)
		var ResponseSigBuf []byte
		if ResponseSigBuf, err = koblitz.SignCompact(koblitz.S256(), &userprivkey, hasher.Sum(nil), false); err != nil {
			b.Fatalf("Error for user signing response: %s", err)
			return
		}
		if len(ResponseSigBuf) != 65 {
			b.Fatalf("Error in test: response signature is not 65 bytes")
			return
		}
		copy(userCommitResponse.CommResponseSig[:], ResponseSigBuf)
		userCommitResponse.PuzzleAnswerReveal = solnorder
		transcript.Responses = append(transcript.Responses, userCommitResponse)
	}

	var fullTScript []byte
	if fullTScript, err = transcript.Serialize(); err != nil {
		b.Fatalf("Error serializing transcript: %s", err)
		return
	}

	b.SetBytes(int64(len(fullTScript)))
	b.StopTimer()
	b.ResetTimer()

	// NOTE: we are ONLY benchmarking verification time
	var valid bool
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		valid, err = transcript.Verify()
		b.StopTimer()

		if !valid {
			b.Fatalf("Empty transcript should have been valid, was invalid: %s", err)
			return
		}
	}

	// b.Logf("Transcript bytes: %d", len(fullTScript))
	// b.Logf("Orders processed: %d", orders)
	// b.Logf("Transcript bytes per user: %f", float64(len(fullTScript))/float64(orders))
	return
}

// TODO: Create benchmark which re-uses SolutionOrders, verify
// completely in parallel

func BenchmarkValidTranscript100M(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptVerify(g, 100000000, amt)
		})
	}
}

func BenchmarkValidTranscript1M(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptVerify(g, 1000000, amt)
		})
	}
}

func BenchmarkValidTranscript10K(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptVerify(g, 10000, amt)
		})
	}
}

// TODO: replace this function it is unnecessary
// I am sorry for introducing this code into existence
func hash256(buf []byte) (sum []byte) {
	sumBacking := [32]byte{}
	hasher := sha3.New256()
	hasher.Write(buf)
	copy(sumBacking[:], hasher.Sum(nil))
	sum = sumBacking[:]
	return
}

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
	// if idsig, err = exprivkey.Sign(hash256(emptyTranscript.BatchId[:])); err != nil {
	// 	err = fmt.Errorf("Error with exchange signing batch id: %s", err)
	// 	return
	// }
	// emptyTranscript.BatchIdSig = idsig.Serialize()
	var batchSig []byte
	if batchSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, hash256(emptyTranscript.BatchId[:]), false); err != nil {
		t.Errorf("Error compact signing batch id sig: %s", err)
		return
	}
	emptyTranscript.BatchIdSig = make([]byte, len(batchSig))
	copy(emptyTranscript.BatchIdSig, batchSig)

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
		if soln, err = NewSolutionOrder(1024); err != nil {
			t.Errorf("Error creating solution order of 1024 bits: %s", err)
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
		emptyTranscript.PuzzledOrders = append(emptyTranscript.PuzzledOrders, signedOrder)
	}

	// now that we have a bunch of puzzled orders, we should create a
	// commitment out of it.
	var CommitmentPreImg []byte
	for _, encOrder := range emptyTranscript.PuzzledOrders {
		var rawPuzzle []byte
		if rawPuzzle, err = encOrder.Serialize(); err != nil {
			t.Errorf("Error serializing submitted order before hashing: %s", err)
			return
		}
		CommitmentPreImg = append(CommitmentPreImg, rawPuzzle...)
	}
	copy(emptyTranscript.Commitment[:], hash256(CommitmentPreImg))
	var exchangeCommSig []byte
	if exchangeCommSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, emptyTranscript.Commitment[:], false); err != nil {
		t.Errorf("Error with exchange signing the commitment: %s", err)
		return
	}
	emptyTranscript.CommitSig = make([]byte, len(exchangeCommSig))
	copy(emptyTranscript.CommitSig, exchangeCommSig)

	// users now create their signatures and reveal solutions
	for userprivkey, solnorder := range privkeyOrderMap {
		// because we're running a test we do not check the time -- in
		// reality you NEED to check the time elapsed.
		userCommitResponse := CommitResponse{}

		// h(commit + sig + answer) = e
		var responseBuf []byte
		responseBuf = append(responseBuf, emptyTranscript.Commitment[:]...)
		responseBuf = append(responseBuf, emptyTranscript.CommitSig...)
		var solnOrderBuf []byte
		if solnOrderBuf, err = solnorder.Serialize(); err != nil {
			t.Errorf("Error serializing solution order for response: %s", err)
			return
		}
		responseBuf = append(responseBuf, solnOrderBuf...)
		var ResponseSigBuf []byte
		if ResponseSigBuf, err = koblitz.SignCompact(koblitz.S256(), &userprivkey, hash256(responseBuf), false); err != nil {
			t.Errorf("Error for user signing response: %s", err)
			return
		}
		if len(ResponseSigBuf) != 65 {
			t.Errorf("Error in test: response signature is not 65 bytes")
			return
		}
		copy(userCommitResponse.CommResponseSig[:], ResponseSigBuf)
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

func generateTranscript(time uint64, orders uint64) (tscript Transcript, err error) {
	if orders == 0 {
		err = fmt.Errorf("Cannot run test with no orders, please setup test correctly")
		return
	}

	// NOTE: we only care about how long it takes to solve
	// create exchange private key
	var exprivkey *koblitz.PrivateKey
	if exprivkey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Error creating exchange private key for signing: %s", err)
		return
	}

	// init empty tscript, the id from there is valid

	// var idsig *koblitz.Signature
	// if idsig, err = exprivkey.Sign(hash256(tscript.BatchId[:])); err != nil {
	// 	err = fmerr = fmt.Errorf("Error with exchange signing batch id: %s", err)
	// 	return
	// }
	// tscript.BatchIdSig = idsig.Serialize()
	hasher := sha3.New256()
	hasher.Write(tscript.BatchId[:])
	var batchSig []byte
	if batchSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, hasher.Sum(nil), false); err != nil {
		err = fmt.Errorf("Error compact signing batch id sig: %s", err)
		return
	}
	tscript.BatchIdSig = make([]byte, len(batchSig))
	copy(tscript.BatchIdSig, batchSig)

	hasher.Reset()
	// This maps private key to solution order so we can respond
	// correctly later.
	var privkeyOrderMap map[koblitz.PrivateKey]SolutionOrder = make(map[koblitz.PrivateKey]SolutionOrder)
	// var bufMtx sync.Mutex
	var solnBuf []SolutionOrder = make([]SolutionOrder, orders)
	var errBuf []error = make([]error, orders)
	var wg sync.WaitGroup
	wg.Add(int(orders))
	for i := uint64(0); i < orders; i++ {
		// First create solution
		solnBuf[i].P = new(big.Int)
		solnBuf[i].Q = new(big.Int)
		go func(j int) {
			var soln SolutionOrder
			if soln, errBuf[j] = NewSolutionOrder(1024); errBuf[j] != nil {
				errBuf[j] = fmt.Errorf("Error creating solution order of 1024 bits: %s", errBuf[j])
				wg.Done()
				return
			}
			solnBuf[j].P.Set(soln.P)
			solnBuf[j].Q.Set(soln.Q)
			wg.Done()
		}(int(i))
	}
	wg.Wait()

	for j, creationErr := range errBuf {
		if creationErr != nil {
			err = fmt.Errorf("Error: the %d'th goroutine failed with error: %s", j, creationErr)
			return
		}
	}

	for _, soln := range solnBuf {
		// NOTE: start of user stuff
		var userPrivKey *koblitz.PrivateKey
		if userPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
			err = fmt.Errorf("Error creating new user private key for signing: %s", err)
			return
		}

		privkeyOrderMap[*userPrivKey] = soln

		// now create encrypted order
		var encOrder EncryptedSolutionOrder
		if encOrder, err = soln.EncryptSolutionOrder(*origOrder, time); err != nil {
			err = fmt.Errorf("Error encrypting solution order for test: %s", err)
			return
		}

		var encOrderBuf []byte
		if encOrderBuf, err = encOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing encrypted order before signing: %s", err)
			return
		}

		hasher.Reset()
		hasher.Write(encOrderBuf)
		var userSigBuf []byte
		if userSigBuf, err = koblitz.SignCompact(koblitz.S256(), userPrivKey, hasher.Sum(nil), false); err != nil {
			err = fmt.Errorf("Error signing encrypted order for user: %s", err)
			return
		}

		// now that we've created the solution order we add it to the
		// tscript as being "submitted".
		signedOrder := SignedEncSolOrder{
			EncSolOrder: encOrder,
			Signature:   make([]byte, len(userSigBuf)),
		}

		copy(signedOrder.Signature, userSigBuf)
		tscript.PuzzledOrders = append(tscript.PuzzledOrders, signedOrder)
	}

	// now that we have a bunch of puzzled orders, we should create a
	// commitment out of it.
	hasher.Reset()
	for _, encOrder := range tscript.PuzzledOrders {
		var rawPuzzle []byte
		if rawPuzzle, err = encOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing submitted order before hashing: %s", err)
			return
		}
		hasher.Write(rawPuzzle)
	}
	copy(tscript.Commitment[:], hasher.Sum(nil))
	var exchangeCommSig []byte
	if exchangeCommSig, err = koblitz.SignCompact(koblitz.S256(), exprivkey, tscript.Commitment[:], false); err != nil {
		err = fmt.Errorf("Error with exchange signing the commitment: %s", err)
		return
	}
	tscript.CommitSig = make([]byte, len(exchangeCommSig))
	copy(tscript.CommitSig, exchangeCommSig)

	hasher.Reset()

	// users now create their signatures and reveal solutions
	for userprivkey, solnorder := range privkeyOrderMap {
		// because we're running a test we do not check the time -- in
		// reality you NEED to check the time elapsed.
		userCommitResponse := CommitResponse{}

		// h(commit + sig + answer) = e
		hasher.Reset()
		hasher.Write(tscript.Commitment[:])
		hasher.Write(tscript.CommitSig)
		var solnOrderBuf []byte
		if solnOrderBuf, err = solnorder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing solution order for response: %s", err)
			return
		}
		hasher.Write(solnOrderBuf)
		var ResponseSigBuf []byte
		if ResponseSigBuf, err = koblitz.SignCompact(koblitz.S256(), &userprivkey, hasher.Sum(nil), false); err != nil {
			err = fmt.Errorf("Error for user signing response: %s", err)
			return
		}
		if len(ResponseSigBuf) != 65 {
			err = fmt.Errorf("Error in test: response signature is not 65 bytes")
			return
		}
		copy(userCommitResponse.CommResponseSig[:], ResponseSigBuf)
		userCommitResponse.PuzzleAnswerReveal = solnorder
		tscript.Responses = append(tscript.Responses, userCommitResponse)
	}
	return
}

// runBenchTranscriptVerify runs a benchmark which creates orders with
// a time parameter specified by the user, and creates a valid
// transcript.
func runBenchTranscriptSolve(b *testing.B, time uint64, orders uint64) {
	var err error
	b.StopTimer()
	b.ResetTimer()

	var transcript Transcript
	if transcript, err = generateTranscript(time, orders); err != nil {
		b.Errorf("Error generating transcript for solve benchmark: %s", err)
		return
	}

	var valid bool
	valid, err = transcript.Verify()

	if !valid {
		b.Fatalf("Empty transcript should have been valid, was invalid: %s", err)
		return
	}

	// NOTE: this is ALL we are tracking!
	var fullTScript []byte
	if fullTScript, err = transcript.Serialize(); err != nil {
		b.Fatalf("Error serializing transcript: %s", err)
		return
	}

	// set the bytes we are processing
	b.SetBytes(int64(len(fullTScript)))

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _, err = transcript.Solve()
	}
	b.StopTimer()

	if err != nil {
		b.Fatalf("Transcript should have been easily solvable, errored instead: %s", err)
		return
	}

	var solved []AuctionOrder
	if solved, _, err = transcript.Solve(); err != nil {
		b.Fatalf("Transcript should have been easily solvable, errored instead: %s", err)
		return
	}
	if len(solved) != int(orders) {
		b.Fatalf("Exchange could not solve all orders, it should have been able to: %s", err)
		return
	}

	return
}

// runBenchTranscriptSerialize runs a benchmark which creates orders with
// a time parameter specified by the user, and benchmarks the time it
// takes to serialize
func runBenchTranscriptSerialize(b *testing.B, time uint64, orders uint64) {
	var err error
	b.StopTimer()
	b.ResetTimer()

	var transcript Transcript
	if transcript, err = generateTranscript(time, orders); err != nil {
		b.Errorf("Error generating transcript for solve benchmark: %s", err)
		return
	}

	var valid bool
	valid, err = transcript.Verify()

	if !valid {
		b.Fatalf("Empty transcript should have been valid, was invalid: %s", err)
		return
	}

	// NOTE: this is ALL we are tracking!
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = transcript.Serialize()
	}
	b.StopTimer()

	if err != nil {
		b.Fatalf("Error serializing transcript: %s", err)
		return
	}

	return
}

func BenchmarkSerializeTranscript(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 80, 100, 200, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptSerialize(g, 10000, amt)
		})
	}
}

func BenchmarkSolveTranscript10K(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptSolve(g, 10000, amt)
		})
	}
}

func BenchmarkSolveTranscript1M(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptSolve(g, 1000000, amt)
		})
	}
}

func BenchmarkSolveTranscript100M(b *testing.B) {
	orderAmounts := []uint64{1, 10, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000, 1500, 2000}
	for _, amt := range orderAmounts {
		b.Run(fmt.Sprintf("NumTranscripts_%d", amt), func(g *testing.B) {
			runBenchTranscriptSolve(g, 100000000, amt)
		})
	}
}
