package cxbenchmark

import (
	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"
)

// createWhitelistMap creates a map from each coin provided to the list of pukeys, in the correct format to be passed into a PinkySwear settlement engine
func createWhitelistMap(coinList []*coinparam.Params, pubkeys []*koblitz.PublicKey) (wlMap map[*coinparam.Params][][33]byte) {
	wlMap = make(map[*coinparam.Params][][33]byte)

	for _, coin := range coinList {
		wlMap[coin] = kobPubToBytes(pubkeys)
	}

	return
}

// kobPubToBytes turns a []*koblitz.PublicKey to a [][33]byte by serializing each key
func kobPubToBytes(pubList []*koblitz.PublicKey) (retList [][33]byte) {
	// We know the length so we can do fun indexing rather than appending all the time
	// and actually preserve order
	retList = make([][33]byte, len(pubList))
	var currSerialized [33]byte
	for i, pubkey := range pubList {
		copy(currSerialized[:], pubkey.SerializeCompressed())
		retList[i] = currSerialized
	}

	return
}
