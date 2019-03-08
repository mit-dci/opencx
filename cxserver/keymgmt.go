package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/btcutil/base58"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/wallit"
)

// NewAddressLTC returns a new address based on the keygen retrieved from the wallet
func (server *OpencxServer) NewAddressLTC(pubkey *koblitz.PublicKey) (string, error) {
	// No really what is this
	return server.getLTCAddrFunc()(pubkey)
}

// NewAddressBTC returns a new address based on the keygen retrieved from the wallet
func (server *OpencxServer) NewAddressBTC(pubkey *koblitz.PublicKey) (string, error) {
	// What is this
	return server.getBTCAddrFunc()(pubkey)
}

// NewAddressVTC returns a new address based on the keygen retrieved from the wallet
func (server *OpencxServer) NewAddressVTC(pubkey *koblitz.PublicKey) (string, error) {
	// Is this currying
	return server.getVTCAddrFunc()(pubkey)
}

// getVTCAddrFunc is used by NewAddressVTC as well as UpdateAddresses to call the address closure
func (server *OpencxServer) getVTCAddrFunc() func(pubkey *koblitz.PublicKey) (string, error) {
	return GetAddrFunction(server.OpencxVTCWallet)
}

// getBTCAddrFunc is used by NewAddressBTC as well as UpdateAddresses to call the address closure
func (server *OpencxServer) getBTCAddrFunc() func(pubkey *koblitz.PublicKey) (string, error) {
	return GetAddrFunction(server.OpencxBTCWallet)
}

// getLTCAddrFunc is used by NewAddressLTC as well as UpdateAddresses to call the address closure
func (server *OpencxServer) getLTCAddrFunc() func(pubkey *koblitz.PublicKey) (string, error) {
	return GetAddrFunction(server.OpencxLTCWallet)
}

// GetAddrFunction returns a function that can safely be called by the DB
func GetAddrFunction(wallet *wallit.Wallit) func(*koblitz.PublicKey) (string, error) {
	pubKeyHashAddrID := wallet.Param.PubKeyHashAddrID
	return func(pubkey *koblitz.PublicKey) (addr string, err error) {
		// TODO: in the future this should be deterministic based on public key.
		// This is to make it really easy to figure out stuff

		defer func() {
			if err != nil {
				err = fmt.Errorf("Problem with address closure: \n%s", err)
			}
		}()

		// Create a new address
		var addrBytes [20]byte
		if addrBytes, err = wallet.NewAdr160(); err != nil {
			return
		}

		// encode it to store in own db
		addr = base58.CheckEncode(addrBytes[:], pubKeyHashAddrID)

		return
	}
}
