package match

// Transcript is the representation of a non front running proof
// transcript. Puzzled orders are the "batch" and this should be able
// to be verified quickly.
type Transcript struct {
	batchId          AuctionID                `json:batchid`
	batchIdSig       []byte                   `json:"signature"`
	puzzledOrders    []EncryptedSolutionOrder `json:"puzzledorders"`
	commitment       [32]byte                 `json:"commitment"`
	commitSig        []byte                   `json:"commitsig"`
	commResponseSigs [][]byte                 `json:"commresponsesigs"`
	decryptedOrders  []SolutionOrder          `json:"decryptedorders"`
	// TODO: finish rest of transcript
}

// Verify verifies the signatures in the transcript and ensures
// that the batch was carried out correctly.
func (tr *Transcript) Verify() (err error) {
	return
}
