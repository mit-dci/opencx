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

// TestSHAAES is the only hash puzzle here because it's the only block cipher that supports
// 256 bit keys, and if we're using the resulting hash as the key then ideally it should be
// a correct key size. RSW is the interesting one anyways.
func TestSHAAES(t *testing.T) {
	message := make([]byte, 32)
	copy(message, []byte("SHA256 hash time puzzle with AES!"))
	// This should take a couple seconds
	ciphertext, puzzle, err := CreateSHAPuzzleAES(1000000, message)
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

func TestRSWRC5ManyN8_T100000(t *testing.T) {
	solveRSWRC5Concurrent(uint64(100000), uint64(8), t)
	return
}

func solveRSWRC5Concurrent(timeToSolve uint64, howMany uint64, t *testing.T) {
	resChan := make(chan bool, howMany)
	for i := uint64(0); i < howMany; i++ {
		message := make([]byte, 32)
		copy(message, []byte("RSW96 Full Scheme!!!!!"))
		// This should take a couple seconds
		ciphertext, puzzle, err := CreateRSW2048A2PuzzleRC5(timeToSolve, message)
		if err != nil {
			t.Fatalf("Error creating puzzle: %s", err)
		}

		go func() {
			newMessage, err := SolvePuzzleRC5(ciphertext, puzzle)
			if err != nil {
				t.Fatalf("Error solving puzzle: %s", err)
			}

			if !bytes.Equal(newMessage, message) {
				t.Fatalf("Messages not equal")
			}

			t.Logf("Solved concurrent")
			resChan <- true
		}()
	}

	for i := uint64(0); i < howMany; i++ {
		<-resChan
	}

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
