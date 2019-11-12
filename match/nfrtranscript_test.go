package match

import (
	"testing"
)

func TestEmptyTranscripVerify(t *testing.T) {
	var err error
	emptyTranscript := Transcript{}

	var valid bool
	if valid, err = emptyTranscript.Verify(); err != nil {
		t.Errorf("Error verifying empty transcript: %s", err)
		return
	}

	if valid {
		t.Errorf("Empty transcript should have been invalid, was valid")
		return
	}
	return
}
