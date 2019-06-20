package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

// DebitUser adds to the balance of the pubkey by issuing a settlement exec and bringing it through
// all of the required data stores.
// DebitUser acquires dbLock so it can just be called.
func (server *OpencxServer) DebitUser(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {

	var assetToDebit match.Asset
	if assetToDebit, err = match.AssetFromCoinParam(param); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for DebitUser: %s", err)
		return
	}

	// Lock!
	server.dbLock.Lock()

	// Get the settle store and the settle engine for the coin
	var currSettleStore cxdb.SettlementStore
	var ok bool
	if currSettleStore, ok = server.SettlementStores[param]; !ok {
		err = fmt.Errorf("Could not find settlement store for cointype %s: %s", param.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleEngine match.SettlementEngine
	if currSettleEngine, ok = server.SettlementEngines[param]; !ok {
		err = fmt.Errorf("Could not find settlement engine for cointype %s: %s", param.Name, err)
		server.dbLock.Unlock()
		return
	}

	setExecForPush := &match.SettlementExecution{
		Type:   match.Debit,
		Asset:  assetToDebit,
		Amount: amount,
	}
	copy(setExecForPush.Pubkey[:], pubkey.SerializeCompressed())

	var valid bool
	if valid, err = currSettleEngine.CheckValid(setExecForPush); err != nil {
		err = fmt.Errorf("Error checking valid exec for DebitUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	var setRes *match.SettlementResult
	var settlementResults []*match.SettlementResult
	if valid {
		if setRes, err = currSettleEngine.ApplySettlementExecution(setExecForPush); err != nil {
			err = fmt.Errorf("Error applying settlement exec for DebitUser: %s", err)
			server.dbLock.Unlock()
			return
		}
	} else {
		err = fmt.Errorf("Error, invalid settlement exec for DebitUser")
		server.dbLock.Unlock()
		return
	}

	settlementResults = append(settlementResults, setRes)

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for DebitUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	server.dbLock.Unlock()
	return
}

// CreditUser subtracts the balance of the pubkey by issuing a settlement exec and bringing it through
// all of the required data stores.
// CreditUser acquires dbLock so it can just be called.
func (server *OpencxServer) CreditUser(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {

	var assetToCredit match.Asset
	if assetToCredit, err = match.AssetFromCoinParam(param); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for CreditUser: %s", err)
		return
	}

	// Lock!
	server.dbLock.Lock()

	// Get the settle store and the settle engine for the coin
	var currSettleStore cxdb.SettlementStore
	var ok bool
	if currSettleStore, ok = server.SettlementStores[param]; !ok {
		err = fmt.Errorf("Could not find settlement store for cointype %s: %s", param.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleEngine match.SettlementEngine
	if currSettleEngine, ok = server.SettlementEngines[param]; !ok {
		err = fmt.Errorf("Could not find settlement engine for cointype %s: %s", param.Name, err)
		server.dbLock.Unlock()
		return
	}

	setExecForPush := &match.SettlementExecution{
		Type:   match.Credit,
		Asset:  assetToCredit,
		Amount: amount,
	}
	copy(setExecForPush.Pubkey[:], pubkey.SerializeCompressed())

	var valid bool
	if valid, err = currSettleEngine.CheckValid(setExecForPush); err != nil {
		err = fmt.Errorf("Error checking valid exec for CreditUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	var setRes *match.SettlementResult
	var settlementResults []*match.SettlementResult
	if valid {
		if setRes, err = currSettleEngine.ApplySettlementExecution(setExecForPush); err != nil {
			err = fmt.Errorf("Error applying settlement exec for CreditUser: %s", err)
			server.dbLock.Unlock()
			return
		}
	} else {
		err = fmt.Errorf("Error, invalid settlement exec for CreditUser")
		server.dbLock.Unlock()
		return
	}
	settlementResults = append(settlementResults, setRes)

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for CreditUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	server.dbLock.Unlock()
	return
}
