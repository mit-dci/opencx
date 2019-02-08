package util

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/mit-dci/lit/btcutil/chaincfg"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"
)

// I should probably test that the neutered private key derived with the same i is equal to the public key

// Keychain is a struct that holds public keys
type Keychain struct {
	BTCPubKey *hdkeychain.ExtendedKey
	LTCPubKey *hdkeychain.ExtendedKey
	VTCPubKey *hdkeychain.ExtendedKey
	BTCParams *coinparam.Params
	LTCParams *coinparam.Params
	VTCParams *coinparam.Params
}

// the trick is to store i in the db along with the pub key so you know how to derive the priv key only if you have the master
// Should I make the child address dependent on the username or should I figure something else out? I trust encryption enough, so I'll make it determined based on name.

// NewAddressBTC makes a new address for the btc testnet based on the username
func (k *Keychain) NewAddressBTC(username string) (string, error) {
	sha := sha256.New()
	sha.Write([]byte(username))
	// TODO: Make this better lol, one of the most annoying things is key assignment and management
	// We mod by 0x80000000 to make sure it's not hardened
	data := binary.BigEndian.Uint32(sha.Sum(nil)[:]) % 0x80000000

	childKey, err := k.BTCPubKey.Child(data)
	if err != nil {
		return "", fmt.Errorf("Error when deriving child public key for new btc address: \n%s", err)
	}

	paramIDHolder := new(chaincfg.Params)
	paramIDHolder.PubKeyHashAddrID = k.BTCParams.PubKeyHashAddrID
	pkHashAddr, err := childKey.Address(paramIDHolder)
	if err != nil {
		return "", err
	}

	return pkHashAddr.String(), nil
}

// NewAddressVTC makes a new address for the btc testnet based on the username
func (k *Keychain) NewAddressVTC(username string) (string, error) {
	sha := sha256.New()
	sha.Write([]byte(username))
	// TODO: Make this better lol, one of the most annoying things is key assignment and management
	// We mod by 0x80000000 to make sure it's not hardened
	data := binary.BigEndian.Uint32(sha.Sum(nil)[:]) % 0x80000000

	childKey, err := k.VTCPubKey.Child(data)
	if err != nil {
		return "", fmt.Errorf("Error when deriving child public key for new vtc address: \n%s", err)
	}

	paramIDHolder := new(chaincfg.Params)
	paramIDHolder.PubKeyHashAddrID = k.VTCParams.PubKeyHashAddrID
	pkHashAddr, err := childKey.Address(paramIDHolder)
	if err != nil {
		return "", err
	}
	return pkHashAddr.String(), nil
}

// NewAddressLTC makes a new address for the btc testnet based on the username
func (k *Keychain) NewAddressLTC(username string) (string, error) {
	sha := sha256.New()
	sha.Write([]byte(username))

	// TODO: Make this better lol, one of the most annoying things is key assignment and management
	// We mod by 0x80000000 to make sure it's not hardened
	data := binary.BigEndian.Uint32(sha.Sum(nil)[:]) % 0x80000000

	childKey, err := k.LTCPubKey.Child(data)
	if err != nil {
		return "", fmt.Errorf("Error when deriving child public key for new ltc address: \n%s", err)
	}

	paramIDHolder := new(chaincfg.Params)
	paramIDHolder.PubKeyHashAddrID = k.LTCParams.PubKeyHashAddrID
	pkHashAddr, err := childKey.Address(paramIDHolder)
	if err != nil {
		return "", err
	}

	return pkHashAddr.String(), nil
}
