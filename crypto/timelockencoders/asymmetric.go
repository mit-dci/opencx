package timelockencoders

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/mit-dci/opencx/crypto"
	"github.com/mit-dci/opencx/crypto/rsw"
)

// CreateRSW2048A2PuzzleRSA creates a RSW timelock puzzle with time t and encrypts the message using RSA. We marshal the key to PKCS1
func CreateRSW2048A2PuzzleRSA(t uint64, message []byte) (ciphertext []byte, puzzle crypto.Puzzle, err error) {
	// Generate private key
	var rsaPrivKey *rsa.PrivateKey
	if rsaPrivKey, err = rsa.GenerateMultiPrimeKey(rand.Reader, 2, 2048); err != nil {
		err = fmt.Errorf("Error generating key to encrypt ciphertext: %s", err)
		return
	}

	// Set up what the puzzle will encrypt (the rsa key, which we marshal using PKCS#1)
	var timelock crypto.Timelock
	if timelock, err = rsw.New2048A2(x509.MarshalPKCS1PrivateKey(rsaPrivKey)); err != nil {
		err = fmt.Errorf("Error creating new rsw timelock for rsa puzzle: %s", err)
		return
	}

	// Set up the puzzle to send
	if puzzle, _, err = timelock.SetupTimelockPuzzle(t); err != nil {
		err = fmt.Errorf("Error setting up timelock while creating rsa puzzle: %s", err)
		return
	}

	// Encrypt the message into the ciphertext so we can then send it out along with the puzzle
	if ciphertext, err = rsa.EncryptPKCS1v15(rand.Reader, &rsaPrivKey.PublicKey, message); err != nil {
		err = fmt.Errorf("Error creating ciphertext for rsa puzzle: %s", err)
		return
	}

	// We've now sent out the puzzle (which is n, a, t, ck)

	return
}

// SolvePuzzleRSAPKCS1 solves the timelock puzzle and decrypts the ciphertext using RSA. We assume the key is in PKCS1 format
func SolvePuzzleRSAPKCS1(ciphertext []byte, puzzle crypto.Puzzle) (message []byte, err error) {
	if puzzle == nil {
		err = fmt.Errorf("Puzzle cannot be nil, what are you solving")
		return
	}

	var key []byte
	if key, err = puzzle.Solve(); err != nil {
		err = fmt.Errorf("Error solving auction puzzle: %s", err)
		return
	}

	var privkey *rsa.PrivateKey
	if privkey, err = x509.ParsePKCS1PrivateKey(key); err != nil {
		err = fmt.Errorf("Error when parsing private key with pkcs#1 encoding: %s", err)
		return
	}

	if message, err = privkey.Decrypt(rand.Reader, ciphertext, nil); err != nil {
		err = fmt.Errorf("Error when decrypting puzzle message: %s", err)
		return
	}

	return
}

// SolvePuzzleECIES solves the timelock puzzle and decrypts the ciphertext using ECIES. We assume the key is an ASN.1 ECPKS
func SolvePuzzleECIES(ciphertext []byte, puzzle crypto.Puzzle) (message []byte, err error) {
	if puzzle == nil {
		err = fmt.Errorf("Puzzle cannot be nil, what are you solving")
		return
	}

	var key []byte
	if key, err = puzzle.Solve(); err != nil {
		err = fmt.Errorf("Error solving auction puzzle: %s", err)
		return
	}

	var privkey *ecdsa.PrivateKey
	if privkey, err = x509.ParseECPrivateKey(key); err != nil {
		err = fmt.Errorf("Error when parsing private key with ECPKS encoding: %s", err)
		return
	}

	var eciesPrivkey *ecies.PrivateKey
	eciesPrivkey = ecies.ImportECDSA(privkey)

	// sharedInfo1 and sharedInfo2 (s1, s2) are not going to be needed (they are nil) because we assume no shared data
	// between the solver and creator of the puzzle.
	if message, err = eciesPrivkey.Decrypt(ciphertext, nil, nil); err != nil {
		err = fmt.Errorf("Error when decrypting puzzle message: %s", err)
		return
	}

	return
}
