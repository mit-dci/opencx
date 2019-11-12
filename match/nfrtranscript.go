package match

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/crypto/rsw"
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
}

// CommitResponse is the commitment response. The sig is the
// puzzleanswerreveal + the commitment + the commitsig
type CommitResponse struct {
	CommResponseSig    []byte        `json:"commresponse"`
	PuzzleAnswerReveal SolutionOrder `json:"puzzleanswer"`
}

// TODO: Make a struct for commitment response - p, q, commitment, and
// a user signature.

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

	// this is a map from N to the order
	var pzMap map[*big.Int]EncryptedSolutionOrder = make(map[*big.Int]EncryptedSolutionOrder)

	// TODO: make map for puzzle inclusion (puzzle answer => bool)
	var bufForCommitment []byte
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
		if responseSig, err = koblitz.ParseSignature(response.CommResponseSig, koblitz.S256()); err != nil {
			err = fmt.Errorf("Error parsing signature in response: %s", err)
			return
		}

		if !responseSig.Verify(e3, exchangePubKey) {
			err = fmt.Errorf("Invalid user response signature: %s", err)
			return
		}
	}

	// TODO: find a way to connect answers to solution orders, maybe
	// make a map of puzzle.serialize() => bool and reconstruct puzzle
	// when you get the solution orders / responses.
	return
}
