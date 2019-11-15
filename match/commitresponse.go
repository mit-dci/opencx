package match

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// CommitResponse is the commitment response. The sig is the
// puzzleanswerreveal + the commitment + the commitsig
type CommitResponse struct {
	CommResponseSig    [65]byte      `json:"commresponse"`
	PuzzleAnswerReveal SolutionOrder `json:"puzzleanswer"`
}

// Serialize uses gob encoding to turn the commit response into bytes.
func (cr *CommitResponse) Serialize() (raw []byte, err error) {
	var b bytes.Buffer

	// register commit response interface
	gob.Register(CommitResponse{})

	// create a new encoder writing to the buffer
	enc := gob.NewEncoder(&b)

	// encode the commit response in the buffer
	if err = enc.Encode(cr); err != nil {
		err = fmt.Errorf("Error encoding commit response: %s", err)
		return
	}

	// Get the bytes from the buffer
	raw = b.Bytes()
	return
}

// Deserialize turns the commit response from bytes into a usable
// struct.
func (cr *CommitResponse) Deserialize(raw []byte) (err error) {
	var b *bytes.Buffer
	b = bytes.NewBuffer(raw)

	// register CommitResponse
	gob.Register(CommitResponse{})

	// create a new decoder writing to the buffer
	dec := gob.NewDecoder(b)

	// decode the commit response in the buffer
	if err = dec.Decode(cr); err != nil {
		err = fmt.Errorf("Error decoding commitresponse: %s", err)
		return
	}

	return
}
