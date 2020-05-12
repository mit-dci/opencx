package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
)

// GetBalance gets the balance for a specific public key and coin.
func (server *OpencxServer) GetBalance(pubkey *koblitz.PublicKey, coin *coinparam.Params) (amount uint64, err error) {

	// First get the settlement store
	// TODO: There should be two locks, one for critical things such as the matching and settlement
	// engine, and another for just the stuff we view. (like the orderbooks and stores)
	server.dbLock.Lock()
	var currSettlementStore cxdb.SettlementStore
	var ok bool
	if currSettlementStore, ok = server.SettlementStores[coin]; !ok {
		err = fmt.Errorf("Cannot find the settlement store for GetBalance")
		server.dbLock.Unlock()
		return
	}

	if amount, err = currSettlementStore.GetBalance(pubkey); err != nil {
		err = fmt.Errorf("Could not get balance for pubkey for GetBalance: %s", err)
		server.dbLock.Unlock()
		return
	}
	server.dbLock.Unlock()

	return
}
