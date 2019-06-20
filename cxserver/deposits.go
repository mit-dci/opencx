package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
)

// RegisterUser gives the user a balance of 0, and gives them deposit addresses. This acquires locks
// so do not try to do it on your own
func (server *OpencxServer) RegisterUser(pubkey *koblitz.PublicKey) (err error) {

	// go through all params in settlement layer
	addrMap := make(map[*coinparam.Params]string)
	server.dbLock.Lock()
	for param, _ := range server.SettlementEngines {
		if addrMap[param], err = server.GetAddrForCoin(param, pubkey); err != nil {
			err = fmt.Errorf("Error getting address for pubkey and coin for RegisterUser: %s", err)
			server.dbLock.Unlock()
			return
		}
	}
	server.dbLock.Unlock()

	// TODO: might not need such a generous use of locks here
	// Should separate deposit/stores from the engines
	var currDepositStore cxdb.DepositStore
	var ok bool
	for param, addr := range addrMap {
		server.dbLock.Lock()
		if currDepositStore, ok = server.DepositStores[param]; !ok {
			err = fmt.Errorf("Could not find deposit store for %s coin", param.Name)
			server.dbLock.Unlock()
			return
		}

		if err = currDepositStore.RegisterUser(pubkey, addr); err != nil {
			err = fmt.Errorf("Error registering user for deposit address for ingestChannelFund: %s", err)
			server.dbLock.Unlock()
			return
		}
		server.dbLock.Unlock()

		if err = server.DebitUser(pubkey, 0, param); err != nil {
			err = fmt.Errorf("Error giving user a balance of zero for RegisterUser: %s", err)
			return
		}
	}

	return
}

// GetDepositAddress gets the deposit address for the pubkey and the param based on the storage
// in the server
func (server *OpencxServer) GetDepositAddress(pubkey *koblitz.PublicKey, coin *coinparam.Params) (address string, err error) {

	server.dbLock.Lock()
	// first get the deposit store for the boi
	var currDepositStore cxdb.DepositStore
	var ok bool
	if currDepositStore, ok = server.DepositStores[coin]; !ok {
		err = fmt.Errorf("Could not find DepositStore for %s for GetDepositAddress", coin.Name)
		server.dbLock.Unlock()
		return
	}

	if address, err = currDepositStore.GetDepositAddress(pubkey); err != nil {
		err = fmt.Errorf("Error getting deposit address from store for GetDepositAddress: %s", err)
		server.dbLock.Unlock()
		return
	}
	server.dbLock.Unlock()

	return
}
