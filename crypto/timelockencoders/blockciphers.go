package timelockencoders

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"

	"github.com/dgryski/go-rc5"
	"github.com/dgryski/go-rc6"
	"github.com/mit-dci/opencx/crypto"
	"github.com/mit-dci/opencx/crypto/hashtimelock"
	"github.com/mit-dci/opencx/crypto/rsw"
)

func createSHAPuzzle(t uint64, key []byte) (puzzle crypto.Puzzle, anskey []byte, err error) {
	// Set up what the puzzle will encrypt
	var timelock crypto.Timelock
	timelock = hashtimelock.New(key, sha256.New())

	// Set up the puzzle to send
	if puzzle, anskey, err = timelock.SetupTimelockPuzzle(t); err != nil {
		err = fmt.Errorf("Error setting up timelock while creating rsw puzzle: %s", err)
		return
	}

	return
}

func createRSWPuzzle(t uint64, key []byte) (puzzle crypto.Puzzle, anskey []byte, err error) {
	// Set up what the puzzle will encrypt
	var timelock crypto.Timelock
	if timelock, err = rsw.New2048A2(key); err != nil {
		err = fmt.Errorf("Error creating new rsw timelock for rsw puzzle: %s", err)
		return
	}

	// Set up the puzzle to send
	if puzzle, anskey, err = timelock.SetupTimelockPuzzle(t); err != nil {
		err = fmt.Errorf("Error setting up timelock while creating rsw puzzle: %s", err)
		return
	}

	return
}

// createRSWPuzzleWithPrimes creates a RSW puzzle with user-defined
// primes p and q to use as the modulus factors.
func createRSWPuzzleWithPrimes(a uint64, t uint64, key []byte, p *big.Int, q *big.Int) (puzzle crypto.Puzzle, anskey []byte, err error) {
	// Set up what the puzzle will encrypt
	var timelock crypto.Timelock
	if timelock, err = rsw.NewTimelockWithPrimes(key, a, p, q); err != nil {
		err = fmt.Errorf("Error creating new rsw timelock for rsw puzzle: %s", err)
		return
	}

	// Set up the puzzle to send
	if puzzle, anskey, err = timelock.SetupTimelockPuzzle(t); err != nil {
		err = fmt.Errorf("Error setting up timelock while creating rsw puzzle: %s", err)
		return
	}

	return
}

// CreateRC5RSWPuzzleWithPrimes creates a RSW timelock puzzle with
// time t and encrypts the message using RC5, given some user-defined
// primes. This returns a struct unlike the other methods here.
func CreateRC5RSWPuzzleWithPrimes(a uint64, t uint64, message []byte, p *big.Int, q *big.Int) (ciphertext []byte, puzzle rsw.PuzzleRSW, err error) {
	// Generate private key
	var key []byte
	if key, err = Generate16ByteKey(rand.Reader); err != nil {
		err = fmt.Errorf("Could not generate rc5 key for puzzle: %s", err)
		return
	}

	var pzInterface crypto.Puzzle
	if pzInterface, key, err = createRSWPuzzleWithPrimes(a, t, key, p, q); err != nil {
		err = fmt.Errorf("Error creating rsw puzzle with primes before encrypting: %s", err)
		return
	}

	var pzSerialized []byte
	if pzSerialized, err = pzInterface.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing puzzle before encryption: %s", err)
		return
	}

	if err = puzzle.Deserialize(pzSerialized); err != nil {
		err = fmt.Errorf("Error deserializing puzzle before encryption: %s", err)
		return
	}

	if len(key) != 16 {
		err = fmt.Errorf("Error with key size")
		return
	}

	// Create the cipher
	var RC5Cipher cipher.Block
	if RC5Cipher, err = rc5.New(key); err != nil {
		err = fmt.Errorf("Error creating cipher for encryption in rc5 puzzle: %s", err)
		return
	}

	// check to make sure we're going to succeed when encrypting
	if len(message) < RC5Cipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Generate random initialization vector
	var iv []byte
	ciphertext = make([]byte, RC5Cipher.BlockSize()+len(message))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	iv = ciphertext[:RC5Cipher.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		err = fmt.Errorf("Error reading random reader for iv: %s", err)
	}

	// Create encrypter
	var cfbEncrypter cipher.Stream
	cfbEncrypter = cipher.NewCFBEncrypter(RC5Cipher, iv)

	// Actually encrypt
	cfbEncrypter.XORKeyStream(ciphertext[RC5Cipher.BlockSize():], message)
	return
}

// CreateRSW2048A2PuzzleRC5 creates a RSW timelock puzzle with time t and encrypts the message using RC5. This is consistent with the scheme described in RSW96.
func CreateRSW2048A2PuzzleRC5(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	return CreatePuzzleRC5(t, message, createRSWPuzzle)
}

// CreatePuzzleRC5 creates a timelock puzzle with time t and encrypts the message using RC5.
func CreatePuzzleRC5(t uint64, message []byte, puzzleCreator func(uint64, []byte) (crypto.Puzzle, []byte, error)) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	// Generate private key
	var key []byte
	if key, err = Generate16ByteKey(rand.Reader); err != nil {
		err = fmt.Errorf("Could not generate rc5 key for puzzle: %s", err)
		return
	}

	if puzzle, key, err = puzzleCreator(t, key); err != nil {
		err = fmt.Errorf("Error while creating timelock puzzle for rc5: %s", err)
		return
	}

	if len(key) != 16 {
		err = fmt.Errorf("Error with key size")
		return
	}

	// Create the cipher
	var RC5Cipher cipher.Block
	if RC5Cipher, err = rc5.New(key); err != nil {
		err = fmt.Errorf("Error creating cipher for encryption in rc5 puzzle: %s", err)
		return
	}

	// check to make sure we're going to succeed when encrypting
	if len(message) < RC5Cipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Generate random initialization vector
	var iv []byte
	ciphertext = make([]byte, RC5Cipher.BlockSize()+len(message))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	iv = ciphertext[:RC5Cipher.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		err = fmt.Errorf("Error reading random reader for iv: %s", err)
	}

	// Create encrypter
	var cfbEncrypter cipher.Stream
	cfbEncrypter = cipher.NewCFBEncrypter(RC5Cipher, iv)

	// Actually encrypt
	cfbEncrypter.XORKeyStream(ciphertext[RC5Cipher.BlockSize():], message)

	// We've sent out the puzzle (which is n, a, t, ck). We've also sent out cm. This is now consistent with RSW96.
	return
}

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

	// check to make sure we're going to succeed when decrypting
	if len(ciphertext) < RC5Cipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Split up IV into ciphertext and IV
	var iv []byte
	iv = ciphertext[:RC5Cipher.BlockSize()]

	// Don't decrypt the IV
	ciphertext = ciphertext[RC5Cipher.BlockSize():]

	// Make decrypter
	var cfbDecrypter cipher.Stream
	cfbDecrypter = cipher.NewCFBDecrypter(RC5Cipher, iv)

	// make message and then decrypt
	message = make([]byte, len(ciphertext))

	cfbDecrypter.XORKeyStream(message, ciphertext)

	return
}

// CreateRSW2048A2PuzzleRC6 creates a RSW timelock puzzle with time t and encrypts the message using RC6.
func CreateRSW2048A2PuzzleRC6(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	return CreatePuzzleRC6(t, message, createRSWPuzzle)
}

// CreatePuzzleRC6 creates a timelock puzzle with time t and encrypts the message using RC6.
func CreatePuzzleRC6(t uint64, message []byte, puzzleCreator func(uint64, []byte) (crypto.Puzzle, []byte, error)) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	// Generate private key
	var key []byte
	if key, err = Generate16ByteKey(rand.Reader); err != nil {
		err = fmt.Errorf("Could not generate rc6 key for puzzle: %s", err)
		return
	}

	if puzzle, key, err = puzzleCreator(t, key); err != nil {
		err = fmt.Errorf("Error while creating timelock puzzle for rc6: %s", err)
		return
	}

	// Create the cipher
	var RC6Cipher cipher.Block
	if RC6Cipher, err = rc6.New(key); err != nil {
		err = fmt.Errorf("Error creating cipher for encryption in rc6 puzzle: %s", err)
		return
	}

	// check to make sure we're going to succeed when encrypting
	if len(message) < RC6Cipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Generate random initialization vector
	var iv []byte
	ciphertext = make([]byte, RC6Cipher.BlockSize()+len(message))

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	iv = ciphertext[:RC6Cipher.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		err = fmt.Errorf("Error reading random reader for iv: %s", err)
	}

	// Create encrypter
	var cfbEncrypter cipher.Stream
	cfbEncrypter = cipher.NewCFBEncrypter(RC6Cipher, iv)

	// Actually encrypt
	cfbEncrypter.XORKeyStream(ciphertext[RC6Cipher.BlockSize():], message)

	// We've sent out the puzzle (which is n, a, t, ck). We've also sent out cm. This is now consistent with RSW96.

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

	// check to make sure we're going to succeed when decrypting
	if len(ciphertext) < RC6Cipher.BlockSize() {
		err = fmt.Errorf("ciphertext less than blocksize, make a bigger ciphertext")
		return
	}

	// Split up IV into ciphertext and IV
	var iv []byte
	iv = ciphertext[:RC6Cipher.BlockSize()]

	// Don't decrypt the IV
	ciphertext = ciphertext[RC6Cipher.BlockSize():]

	// Make decrypter
	var cfbDecrypter cipher.Stream
	cfbDecrypter = cipher.NewCFBDecrypter(RC6Cipher, iv)

	// make message and then decrypt
	message = make([]byte, len(ciphertext))

	cfbDecrypter.XORKeyStream(message, ciphertext)

	return
}

// CreateSHAPuzzleAES creates a hash timelock puzzle with time t and encrypts the message using AES.
func CreateSHAPuzzleAES(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	return CreatePuzzleAES(t, message, createSHAPuzzle)
}

// CreateRSW2048A2PuzzleAES creates a RSW timelock puzzle with time t and encrypts the message using AES.
func CreateRSW2048A2PuzzleAES(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	return CreatePuzzleAES(t, message, createRSWPuzzle)
}

// CreatePuzzleAES creates a RSW timelock puzzle with time t and encrypts the message using AES.
func CreatePuzzleAES(t uint64, message []byte, puzzleCreator func(uint64, []byte) (crypto.Puzzle, []byte, error)) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	// Generate private key
	var key []byte
	if key, err = Generate16ByteKey(rand.Reader); err != nil {
		err = fmt.Errorf("Could not generate key for aes puzzle: %s", err)
		return
	}

	if puzzle, key, err = puzzleCreator(t, key); err != nil {
		err = fmt.Errorf("Error while creating timelock puzzle for aes: %s", err)
		return
	}

	// Create the cipher
	var AESCipher cipher.Block
	if AESCipher, err = aes.NewCipher(key); err != nil {
		err = fmt.Errorf("Error creating cipher for encryption in aes puzzle: %s", err)
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

// Generate16ByteKey generates a 16 byte long key to be used for AES, RC5, or RC6 from a reader
func Generate16ByteKey(rand io.Reader) (key []byte, err error) {
	key = make([]byte, 16)
	var newKey []byte = make([]byte, 16)
	if _, err = rand.Read(newKey); err != nil {
		err = fmt.Errorf("Error reading from random while generating AES key: %s", err)
		return
	}
	// make sure that the length is 16
	copy(key, newKey)
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
