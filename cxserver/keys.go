package cxserver

import (
	"fmt"

	"github.com/Rjected/lit/btcutil/hdkeychain"
	"github.com/Rjected/lit/coinparam"
)

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(privkey *[32]byte) (err error) {

	// for all settlement engines that we have, make keys
	for param, _ := range server.SettlementEngines {
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
	server.privKeyMtx.Lock()
	server.PrivKeyMap[param] = rootKey
	server.privKeyMtx.Unlock()

	return
}
