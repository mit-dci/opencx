package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"
)

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(privkey *[32]byte) (err error) {

	if err = server.SetupManyKeys(privkey, server.CoinList); err != nil {
		return
	}

	return nil
}

// SetupManyKeys sets up many keys for the server based on an array of coinparams.
func (server *OpencxServer) SetupManyKeys(privkey *[32]byte, paramList []*coinparam.Params) (err error) {
	for _, param := range paramList {
		if err = server.SetupSingleKey(privkey, param); err != nil {
			return
		}
	}
	return
}

// SetupSingleKey sets up a single key based on a single param for the server.
func (server *OpencxServer) SetupSingleKey(privkey *[32]byte, param *coinparam.Params) (err error) {
	var rootKey *hdkeychain.ExtendedKey
	if rootKey, err = hdkeychain.NewMaster(privkey[:], param); err != nil {
		err = fmt.Errorf("Error creating master %s key from private key: \n%s", param.Name, err)
		return
	}
	server.PrivKeyMap[param] = rootKey

	return
}
