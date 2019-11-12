package match

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/crypto/rsw"
	"github.com/mit-dci/opencx/crypto/timelockencoders"
	"golang.org/x/crypto/sha3"
)

// Transcript is the representation of a non front running proof
// transcript. Puzzled orders are the "batch" and this should be able
// to be verified quickly.
type Transcript struct {
	batchId       AuctionID                `json:batchid`
	batchIdSig    []byte                   `json:"signature"`
	puzzledOrders []EncryptedSolutionOrder `json:"puzzledorders"`
	commitment    [32]byte                 `json:"commitment"`
	commitSig     []byte                   `json:"commitsig"`
	responses     []CommitResponse         `json:"responses"`
	solutions     []AuctionOrder           `json:"solutions"`
}

// CommitResponse is the commitment response. The sig is the
// puzzleanswerreveal + the commitment + the commitsig
type CommitResponse struct {
	CommResponseSig    [71]byte      `json:"commresponse"`
	PuzzleAnswerReveal SolutionOrder `json:"puzzleanswer"`
}

// Verify verifies the signatures in the transcript and ensures
// that the batch was carried out correctly.
func (tr *Transcript) Verify() (valid bool, err error) {
	// First verify batch ID
	hasher := sha3.New256()
	if _, err = hasher.Write(tr.batchId[:]); err != nil {
		err = fmt.Errorf("Error writing batch id to hasher: %s", err)
		return
	}

	// e is the hash of the batch ID
	e := hasher.Sum(nil)

	var exchangePubKey *koblitz.PublicKey
	var batchSigValid bool
	if exchangePubKey, batchSigValid, err = koblitz.RecoverCompact(koblitz.S256(), tr.batchIdSig, e); err != nil {
		err = fmt.Errorf("Error recovering pubkey from batch sig: %s", err)
		return
	}

	if !batchSigValid {
		err = fmt.Errorf("Batch id signature invalid: %s", err)
		return
	}

	// TODO: make map for puzzle inclusion (puzzle answer => bool)
	var bufForCommitment []byte
	for _, pzOrder := range tr.puzzledOrders {
		var pzBuf []byte
		if pzBuf, err = pzOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing puzzle order for transcript verification: %s", err)
			return
		}
		bufForCommitment = append(bufForCommitment, pzBuf...)
	}

	hasher = sha3.New256()
	if _, err = hasher.Write(bufForCommitment); err != nil {
		err = fmt.Errorf("Error writing puzzles for commitment to hasher: %s", err)
		return
	}

	// hash of the puzzled orders
	e2 := hasher.Sum(nil)

	var exsig *koblitz.Signature
	if exsig, err = koblitz.ParseSignature(tr.commitSig, koblitz.S256()); err != nil {
		err = fmt.Errorf("Error parsing commitment signature: %s", err)
		return
	}

	if !exsig.Verify(e2, exchangePubKey) {
		err = fmt.Errorf("Invalid commitment signature from exchange")
		return
	}

	if bytes.Compare(e2, tr.commitment[:]) != 0 {
		err = fmt.Errorf("Commitment is not equal to hash of orders - invalid transcript")
		return
	}

	for _, response := range tr.responses {
		// comm + sig + answer = e
		responseHasher := sha3.New256()
		if _, err = responseHasher.Write(tr.commitment[:]); err != nil {
			err = fmt.Errorf("Error writing commitment to hasher: %s", err)
			return
		}
		if _, err = responseHasher.Write(tr.commitSig); err != nil {
			err = fmt.Errorf("Error writing commit sig to hasher: %s", err)
			return
		}

		var answerBytes []byte
		if answerBytes, err = response.PuzzleAnswerReveal.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing answer: %s", err)
			return
		}
		if _, err = responseHasher.Write(answerBytes); err != nil {
			err = fmt.Errorf("Error writing answer bytes to hasher: %s", err)
			return
		}

		e3 := responseHasher.Sum(nil)
		var responseSig *koblitz.Signature
		if responseSig, err = koblitz.ParseSignature(response.CommResponseSig[:], koblitz.S256()); err != nil {
			err = fmt.Errorf("Error parsing signature in response: %s", err)
			return
		}

		if !responseSig.Verify(e3, exchangePubKey) {
			err = fmt.Errorf("Invalid user response signature: %s", err)
			return
		}
	}

	return
}

// Solve processes the encrypted solution orders and the commitment
// responses to partition the encrypted orders into those solvable by
// responses and those that are unsolvable.
func (tr *Transcript) Solve() (solvedOrders []AuctionOrder, invalidResponses []CommitResponse, err error) {
	// this is a map from N to the order
	var pzMap map[*big.Int]EncryptedSolutionOrder = make(map[*big.Int]EncryptedSolutionOrder)
	for _, pzOrder := range tr.puzzledOrders {
		var pzBuf []byte
		if pzBuf, err = pzOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing puzzle order for transcript verification: %s", err)
			return
		}

		var rswPz rsw.PuzzleRSW
		if err = rswPz.Deserialize(pzBuf); err != nil {
			err = fmt.Errorf("Error deserializing puzzle into rsw, ensure that the puzzle sent is valid: %s", err)
			return
		}

		pzMap[rswPz.N] = pzOrder
	}

	var solutionMap map[CommitResponse]EncryptedSolutionOrder = make(map[CommitResponse]EncryptedSolutionOrder)
	var currEnc EncryptedSolutionOrder
	var ok bool
	for _, answer := range tr.responses {
		ok = false
		tempN := new(big.Int)
		tempN.Mul(answer.PuzzleAnswerReveal.p, answer.PuzzleAnswerReveal.q)
		if currEnc, ok = pzMap[tempN]; ok {
			solutionMap[answer] = currEnc
		} else {
			invalidResponses = append(invalidResponses, answer)
		}
	}

	for response, encOrder := range solutionMap {
		var currAuctionOrder AuctionOrder
		if currAuctionOrder, err = trapdoor(response.PuzzleAnswerReveal.p, response.PuzzleAnswerReveal.q, encOrder); err != nil {
			err = fmt.Errorf("Error running trapdoor for revealed answer: %s", err)
			return
		}
		solvedOrders = append(solvedOrders, currAuctionOrder)
	}
	return
}

// calculate trapdoor to solve puzzle
func trapdoor(p, q *big.Int, encOrder EncryptedSolutionOrder) (order AuctionOrder, err error) {
	// calculate trapdoor e = 2^t mod phi(n) = 2^t mod (p-1)(q-1)
	two := big.NewInt(2)
	one := big.NewInt(1)
	pminusone := new(big.Int).Sub(p, one)
	qminusone := new(big.Int).Sub(q, one)
	phi := new(big.Int).Mul(pminusone, qminusone)
	e := new(big.Int).Exp(two, encOrder.OrderPuzzle.T, phi)

	// calculate b = a^e mod N
	b := new(big.Int).Exp(encOrder.OrderPuzzle.A, e, encOrder.OrderPuzzle.N)

	// now b xor c_k = k
	k := new(big.Int).Xor(b, encOrder.OrderPuzzle.CK)
	kBytes := k.Bytes()

	var key []byte
	if len(kBytes) <= 16 {
		key = make([]byte, 16)
	} else {
		key = make([]byte, len(kBytes))
	}
	copy(key, kBytes)

	var orderBytes []byte
	if orderBytes, err = timelockencoders.DecryptPuzzleRC5(encOrder.OrderCiphertext, key); err != nil {
		err = fmt.Errorf("Error decrypting rc5 puzzle from trapdoor key: %s", err)
		return
	}

	if err = order.Deserialize(orderBytes); err != nil {
		err = fmt.Errorf("Error deserializing order for trapdoor into struct: %s", err)
		return
	}
	return
}
