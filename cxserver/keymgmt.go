package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/btcutil/base58"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/wallit"
)

// GetAddrForCoin gets an address based on a wallet and pubkey
func (server *OpencxServer) GetAddrForCoin(coinType *coinparam.Params, pubkey *koblitz.PublicKey) (addr string, err error) {
	wallet, found := server.WalletMap[coinType]
	if !found {
		err = fmt.Errorf("Could not find wallet to create address for")
	}
	if addr, err = GetAddrFunction(wallet)(pubkey); err != nil {
		return
	}

	return
}

// TODO: honestly just delete this at some point. If anyone wants a free pull request just
// make GetAddrFunction a function with 2 parameters.

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
