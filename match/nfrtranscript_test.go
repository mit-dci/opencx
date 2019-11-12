package match

import (
	"testing"
)

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
