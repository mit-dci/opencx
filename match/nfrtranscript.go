package match

type Transcript struct {
	batchId    AuctionID `json:batchid`
	batchIdSig []byte    `json:"signature"`
	// TODO: finish rest of transcript
}

// Verify verifies the signatures in the transcript and ensures
// that the batch was carried out correctly.
func (tr *Transcript) Verify() (err error) {
	return
}
