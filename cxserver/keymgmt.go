package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/btcutil/base58"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
)

// GetAddrForCoin gets an address based on a wallet and pubkey
func (server *OpencxServer) GetAddrForCoin(coinType *coinparam.Params, pubkey *koblitz.PublicKey) (addr string, err error) {
	wallet, found := server.WalletMap[coinType]
	if !found {
		err = fmt.Errorf("Could not find wallet to create address for")
	}

	pubKeyHashAddrID := wallet.Param.PubKeyHashAddrID

	// Create a new address
	var addrBytes [20]byte
	if addrBytes, err = wallet.NewAdr160(); err != nil {
		return
	}

	// encode it to store in own db
	addr = base58.CheckEncode(addrBytes[:], pubKeyHashAddrID)

	return
}
