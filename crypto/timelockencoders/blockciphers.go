package timelockencoders

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/mit-dci/opencx/crypto"
	"github.com/mit-dci/opencx/crypto/rsw"
)

// CreateRSW2048A2PuzzleAES creates a RSW timelock puzzle with time t and encrypts the message using AES.
func CreateRSW2048A2PuzzleAES(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	// Generate private key
	var key []byte
	if key, err = GenerateAESKey(rand.Reader); err != nil {
		err = fmt.Errorf("Could not generate rc5 key for puzzle: %s", err)
		return
	}

	// Set up what the puzzle will encrypt
	var timelock crypto.Timelock
	if timelock, err = rsw.New2048A2(key); err != nil {
		err = fmt.Errorf("Error creating new rsw timelock for aes puzzle: %s", err)
		return
	}

	// Set up the puzzle to send
	if puzzle, _, err = timelock.SetupTimelockPuzzle(t); err != nil {
		err = fmt.Errorf("Error setting up timelock while creating aes puzzle: %s", err)
		return
	}

	// Create the cipher
	var AESCipher cipher.Block
	if AESCipher, err = aes.NewCipher(key); err != nil {
		err = fmt.Errorf("Error creating aes cipher for encryption in puzzle: %s", err)
		return
	}

	// check to make sure we're going to succeed when encrypting
	if len(message) < AESCipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Generate random initialization vector
	var iv []byte
	ciphertext = make([]byte, AESCipher.BlockSize()+len(message))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	iv = ciphertext[:AESCipher.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		err = fmt.Errorf("Error reading random reader for iv: %s", err)
	}

	// Create encrypter
	var cfbEncrypter cipher.Stream
	cfbEncrypter = cipher.NewCFBEncrypter(AESCipher, iv)

	// Actually encrypt
	cfbEncrypter.XORKeyStream(ciphertext[AESCipher.BlockSize():], message)

	// We've sent out the puzzle (which is n, a, t, ck). We've also sent out cm.

	return
}

// GenerateAESKey generates a 16 byte long key to be used for AES from a reader
func GenerateAESKey(rand io.Reader) (key []byte, err error) {
	key = make([]byte, 16)
	if _, err = rand.Read(key); err != nil {
		err = fmt.Errorf("Error reading from random while generating AES key: %s", err)
		return
	}
	return
}

// GenerateIV generates an initialization vector with the block size of the cipher
func GenerateIV(rand io.Reader, blockCipher cipher.Block) (iv []byte, err error) {
	iv = make([]byte, blockCipher.BlockSize())
	if _, err = rand.Read(iv); err != nil {
		err = fmt.Errorf("Error reading from random while generating AES initialization vector: %s", err)
		return
	}
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

	// check to make sure we're going to succeed when decrypting
	if len(ciphertext) < AESCipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Split up IV into ciphertext and IV
	var iv []byte
	iv = ciphertext[:AESCipher.BlockSize()]

	// Don't decrypt the IV
	ciphertext = ciphertext[AESCipher.BlockSize():]

	// Make decrypter
	var cfbDecrypter cipher.Stream
	cfbDecrypter = cipher.NewCFBDecrypter(AESCipher, iv)

	// make message and then decrypt
	message = make([]byte, len(ciphertext))

	cfbDecrypter.XORKeyStream(message, ciphertext)

	return
}
