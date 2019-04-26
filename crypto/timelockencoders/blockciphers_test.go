package timelockencoders

import (
	"bytes"
	"testing"
)

func TestRSWAES(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("RSW96 Full Scheme but with AES!"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateRSW2048A2PuzzleAES(1000000, message)
	if err != nil {
		t.Fatalf("Error creating puzzle: %s", err)
	}

	newMessage, err := SolvePuzzleAES(ciphertext, puzzle)
	if err != nil {
		t.Fatalf("Error solving puzzle: %s", err)
	}

	if !bytes.Equal(newMessage, message) {
		t.Fatalf("Messages not equal")
	}

	t.Logf("We got message: %s!", newMessage)

	return
}

func TestRSWRC5(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("RSW96 Full Scheme!!!!!"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateRSW2048A2PuzzleRC5(1000000, message)
	if err != nil {
		t.Fatalf("Error creating puzzle: %s", err)
	}

	newMessage, err := SolvePuzzleRC5(ciphertext, puzzle)
	if err != nil {
		t.Fatalf("Error solving puzzle: %s", err)
	}

	if !bytes.Equal(newMessage, message) {
		t.Fatalf("Messages not equal")
	}

	t.Logf("We got message: %s!", newMessage)

	return
}

func TestRSWRC6(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("RSW96 Full Scheme but with RC6!"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateRSW2048A2PuzzleRC6(1000000, message)
	if err != nil {
		t.Fatalf("Error creating puzzle: %s", err)
	}

	newMessage, err := SolvePuzzleRC6(ciphertext, puzzle)
	if err != nil {
		t.Fatalf("Error solving puzzle: %s", err)
	}

	if !bytes.Equal(newMessage, message) {
		t.Fatalf("Messages not equal")
	}

	t.Logf("We got message: %s!", newMessage)

	return
}
