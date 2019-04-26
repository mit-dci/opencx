package timelockencoders

import (
	"bytes"
	"testing"
)

func TestRSWRSA(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("RSW96 Full Scheme but with RSA"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateRSW2048A2PuzzleRSA(1000000, message)
	if err != nil {
		t.Fatalf("Error creating puzzle: %s", err)
	}

	newMessage, err := SolvePuzzleRSA(ciphertext, puzzle)
	if err != nil {
		t.Fatalf("Error solving puzzle: %s", err)
	}

	if !bytes.Equal(newMessage, message) {
		t.Fatalf("Messages not equal")
	}

	t.Logf("We got message: %s!", newMessage)

	return
}

func TestRSWECIES(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("RSW96 Full Scheme but with ECIES"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateRSW2048A2PuzzleECIES(1000000, message)
	if err != nil {
		t.Fatalf("Error creating puzzle: %s", err)
	}

	newMessage, err := SolvePuzzleECIES(ciphertext, puzzle)
	if err != nil {
		t.Fatalf("Error solving puzzle: %s", err)
	}

	if !bytes.Equal(newMessage, message) {
		t.Fatalf("Messages not equal")
	}

	t.Logf("We got message: %s!", newMessage)

	return
}
