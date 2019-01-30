package util

import (
	"errors"

	"github.com/mit-dci/lit/btcutil/base58"
	"github.com/mit-dci/lit/coinparam"
	"golang.org/x/crypto/ripemd160"
)

// NewAddressPubKeyHash only exists because the other functions turning pkhashes into bytes don't support coinparam params
func NewAddressPubKeyHash(pkHash []byte, coinType *coinparam.Params) (string, error) {
	return newAddressPubKeyHash(pkHash, coinType.PubKeyHashAddrID)
}

// newAddressPubKeyHash is the internal API to create a pubkey hash address
// with a known leading identifier byte for a network, rather than looking
// it up through its parameters.  This is useful when creating a new address
// structure from a string encoding where the identifer byte is already
// known.
func newAddressPubKeyHash(pkHash []byte, netID byte) (string, error) {
	// Check for a valid pubkey hash length.
	if len(pkHash) != ripemd160.Size {
		return "", errors.New("pkHash must be 20 bytes")
	}

	return base58.CheckEncode(pkHash, netID), nil
}
