package timelocksolvers

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"

	"github.com/dgryski/go-rc5"
	"github.com/dgryski/go-rc6"
	"github.com/mit-dci/opencx/crypto"
)

// SolvePuzzleRC5 solves the timelock puzzle and decrypts the ciphertext using RC5
func SolvePuzzleRC5(ciphertext []byte, puzzle crypto.Puzzle) (message []byte, err error) {
	if puzzle == nil {
		err = fmt.Errorf("Puzzle cannot be nil, what are you solving")
		return
	}

	var key []byte
	if key, err = puzzle.Solve(); err != nil {
		err = fmt.Errorf("Error solving auction puzzle: %s", err)
		return
	}

	var RC5Cipher cipher.Block
	if RC5Cipher, err = rc5.New(key); err != nil {
		err = fmt.Errorf("Could not create new rc5 cipher for puzzle: %s", err)
		return
	}

	// Decrypt doesn't return anything
	RC5Cipher.Decrypt(message, ciphertext)

	return
}

// SolvePuzzleRC6 solves the timelock puzzle and decrypts the ciphertext using RC6
func SolvePuzzleRC6(ciphertext []byte, puzzle crypto.Puzzle) (message []byte, err error) {
	if puzzle == nil {
		err = fmt.Errorf("Puzzle cannot be nil, what are you solving")
		return
	}

	var key []byte
	if key, err = puzzle.Solve(); err != nil {
		err = fmt.Errorf("Error solving auction puzzle: %s", err)
		return
	}

	var RC6Cipher cipher.Block
	if RC6Cipher, err = rc6.New(key); err != nil {
		err = fmt.Errorf("Could not create new rc6 cipher for puzzle: %s", err)
		return
	}

	// Decrypt doesn't return anything
	RC6Cipher.Decrypt(message, ciphertext)

	return
}

// SolvePuzzleAES solves the timelock puzzle and decrypts the ciphertext using AES
func SolvePuzzleAES(ciphertext []byte, puzzle crypto.Puzzle) (message []byte, err error) {
	if puzzle == nil {
		err = fmt.Errorf("Puzzle cannot be nil, what are you solving")
		return
	}

	var key []byte
	if key, err = puzzle.Solve(); err != nil {
		err = fmt.Errorf("Error solving auction puzzle: %s", err)
		return
	}

	var AESCipher cipher.Block
	if AESCipher, err = aes.NewCipher(key); err != nil {
		err = fmt.Errorf("Could not create new aes cipher for puzzle: %s", err)
		return
	}

	// Decrypt doesn't return anything
	AESCipher.Decrypt(message, ciphertext)

	return
}
