package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/qln"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"github.com/mit-dci/opencx/util"
)

// IngestTransactionListAndHeight processes a transaction list and corresponding height
func (server *OpencxServer) ingestTransactionListAndHeight(txList []*wire.MsgTx, height uint64, coinType *coinparam.Params) (err error) {
	// get list of addresses we own
	// check the sender, amounts, receiver of all the transactions
	// check if the receiver is us
	// if so, add the deposit to the table, create a # of confirmations past the height at which it was received

	server.dbLock.Lock()
	// First get the correct deposit store, settlement engine, and settlement store
	var currDepositStore cxdb.DepositStore
	var ok bool
	if currDepositStore, ok = server.DepositStores[coinType]; !ok {
		err = fmt.Errorf("Could not find deposit store for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleStore cxdb.SettlementStore
	if currSettleStore, ok = server.SettlementStores[coinType]; !ok {
		err = fmt.Errorf("Could not find settlement store for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleEngine match.SettlementEngine
	if currSettleEngine, ok = server.SettlementEngines[coinType]; !ok {
		err = fmt.Errorf("Could not find settlement engine for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}

	var addressesWeOwn map[string]*koblitz.PublicKey
	if addressesWeOwn, err = currDepositStore.GetDepositAddressMap(); err != nil {
		// if errors out, unlock
		err = fmt.Errorf("Error getting deposit address map: %s", err)
		server.dbLock.Unlock()
		return
	}
	server.dbLock.Unlock()

	var deposits []match.Deposit

	for _, tx := range txList {
		for _, output := range tx.TxOut {

			// figure out if it's the right script type
			scriptType, data := util.ScriptType(output.PkScript)
			if scriptType == "P2PKH" {
				// It's P2PKH, let's get the address
				var addr *btcutil.AddressPubKeyHash
				if addr, err = btcutil.NewAddressPubKeyHash(data, coinType); err != nil {
					err = fmt.Errorf("Error deciding p2pkh while ingesting txs: %s", err)
					return
				}

				if pubkey, found := addressesWeOwn[addr.String()]; found {
					newDeposit := match.Deposit{
						Pubkey:              pubkey,
						Address:             addr.String(),
						Amount:              uint64(output.Value),
						Txid:                tx.TxHash().String(),
						CoinType:            coinType,
						BlockHeightReceived: height,
						Confirmations:       6,
					}

					logging.Infof("Received deposit for %d %s", newDeposit.Amount, newDeposit.CoinType.Name)
					logging.Infof("%s\n", newDeposit.String())
					deposits = append(deposits, newDeposit)
				}
			}
		}
	}

	if err = server.updateDepositsAtHeight(deposits, height, coinType); err != nil {
		err = fmt.Errorf("Error updating deposits at height for ingestTransactionListAndHeight: %s", err)
		return
	}

	logging.Debugf("Finished ingesting %s block at height %d", coinType.Name, height)
	if height%10000 == 0 {
		logging.Infof("Finished ingesting %s block at height %d\n", coinType.Name, height)
	}
	return
}

// ingestChannelPush changes the user's balance to reflect that a push on a channel happened
func (server *OpencxServer) ingestChannelPush(pushAmt uint64, pubkey *koblitz.PublicKey, coinType uint32) (err error) {

	logging.Infof("Confirmed push from %x to give me %d of cointype %d\n", pubkey.SerializeCompressed(), pushAmt, coinType)
	var param *coinparam.Params
	if param, err = util.GetParamFromHDCoinType(coinType); err != nil {
		err = fmt.Errorf("Error getting param from cointype while ingesting channel push: %s", err)
		return
	}

	if err = server.debitUser(pubkey, pushAmt, param); err != nil {
		err = fmt.Errorf("Error debiting user for ingestChannelPush: %s", err)
		return
	}

	return
}

// ingestChannelConfirm changes the user's balance to reflect that a confirmation of a channel happened
func (server *OpencxServer) ingestChannelConfirm(state *qln.StatCom, pubkey *koblitz.PublicKey, coinType uint32) (err error) {
	logging.Infof("Confirmed channel from %x to give me %d of cointype %d\n", pubkey.SerializeCompressed(), state.MyAmt, coinType)

	var param *coinparam.Params
	if param, err = util.GetParamFromHDCoinType(coinType); err != nil {
		err = fmt.Errorf("Error getting param from hdcointype for ingestChannelConfirm")
		return
	}

	if err = server.debitUser(pubkey, uint64(state.MyAmt), param); err != nil {
		err = fmt.Errorf("Error debiting user for ingestChannelConfirm: %s", err)
		return
	}

	logging.Infof("Confirmed channel from pubkey %x\n", pubkey.SerializeCompressed())

	return
}

// updateDepositsAtHeight acquires locks and does all of the required actions to update the exchange
// when deposits come in at a certain block for a certain coin
func (server *OpencxServer) updateDepositsAtHeight(deposits []match.Deposit, height uint64, coinType *coinparam.Params) (err error) {
	server.dbLock.Lock()
	// First get the correct deposit store, settlement engine, and settlement store
	var currDepositStore cxdb.DepositStore
	var ok bool
	if currDepositStore, ok = server.DepositStores[coinType]; !ok {
		err = fmt.Errorf("Could not find deposit store for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleStore cxdb.SettlementStore
	if currSettleStore, ok = server.SettlementStores[coinType]; !ok {
		err = fmt.Errorf("Could not find settlement store for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}

	var currSettleEngine match.SettlementEngine
	if currSettleEngine, ok = server.SettlementEngines[coinType]; !ok {
		err = fmt.Errorf("Could not find settlement engine for cointype %s: %s", coinType.Name, err)
		server.dbLock.Unlock()
		return
	}
	var depositExecs []*match.SettlementExecution
	if depositExecs, err = currDepositStore.UpdateDeposits(deposits, height); err != nil {
		// if errors out, unlock
		err = fmt.Errorf("Error updating deposits for updateDepositsAtHeight: %s", err)
		server.dbLock.Unlock()
		return
	}

	var settlementResults []*match.SettlementResult
	for _, setExec := range depositExecs {
		// We always check validity first
		var valid bool
		if valid, err = currSettleEngine.CheckValid(setExec); err != nil {
			err = fmt.Errorf("Error checking exec validity for updateDepositsAtHeight: %s", err)
			server.dbLock.Unlock()
			return
		}

		if valid {
			var setRes *match.SettlementResult
			if setRes, err = currSettleEngine.ApplySettlementExecution(setExec); err != nil {
				err = fmt.Errorf("Error applying settlement exec for updateDepositsAtHeight: %s", err)
				server.dbLock.Unlock()
				return
			}
			settlementResults = append(settlementResults, setRes)
		} else {
			err = fmt.Errorf("Error, invalid settlement exec for updateDepositsAtHeight")
			server.dbLock.Unlock()
			return
		}
	}

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for updateDepositsAtHeight: %s", err)
		server.dbLock.Unlock()
		return
	}
	server.dbLock.Unlock()
	return
}

// debutUser adds to the balance of the pubkey by issuing a settlement exec and bringing it through
// all of the required data stores.
// debitUser acquires dbLock so it can just be called.
func (server *OpencxServer) debitUser(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {

	var assetToDebit match.Asset
	if assetToDebit, err = match.AssetFromCoinParam(param); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for debitUser: %s", err)
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

	var setRes *match.SettlementResult
	var valid bool
	if valid, err = currSettleEngine.CheckValid(setExecForPush); err != nil {
		err = fmt.Errorf("Error checking valid exec for debitUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	var settlementResults []*match.SettlementResult
	if valid {
		var setRes *match.SettlementResult
		if setRes, err = currSettleEngine.ApplySettlementExecution(setExecForPush); err != nil {
			err = fmt.Errorf("Error applying settlement exec for debitUser: %s", err)
			server.dbLock.Unlock()
			return
		}
		settlementResults = append(settlementResults, setRes)
	} else {
		err = fmt.Errorf("Error, invalid settlement exec for debitUser")
		server.dbLock.Unlock()
		return
	}

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for debitUser: %s", err)
		server.dbLock.Unlock()
		return
	}

	server.dbLock.Unlock()
	return
}

// ingestChannelFund only registers the user for deposit addresses because the fund channel hasn't
// necessarily been confirmed yet
func (server *OpencxServer) ingestChannelFund(state *qln.StatCom, pubkey *koblitz.PublicKey, coinType uint32, qchanID uint32) (err error) {
	logging.Infof("Pubkey %x funded a channel to give me %d of cointype %d\n", pubkey.SerializeCompressed(), state.MyAmt, coinType)

	var addrMap map[*coinparam.Params]string
	if addrMap, err = server.GetAddressMap(pubkey); err != nil {
		err = fmt.Errorf("Error getting address map for pubkey for ingestChannelFund: %s", err)
		return
	}

	logging.Infof("Registering user with pubkey %x\n", pubkey.SerializeCompressed())

	server.dbLock.Lock()
	var currDepositStore cxdb.DepositStore
	var ok bool
	for param, addr := range addrMap {
		if currDepositStore, ok = server.DepositStores[param]; !ok {
			err = fmt.Errorf("Could not find deposit store for %s coin", param.Name)
			return
		}

		if err = currDepositStore.RegisterUser(pubkey, addr); err != nil {
			err = fmt.Errorf("Error registering user for deposit address for ingestChannelFund: %s", err)
			return
		}
	}
	server.dbLock.Unlock()

	if err = server.SetupFundBack(pubkey, coinType, server.defaultCapacity); err != nil {
		err = fmt.Errorf("Error setting up fund back for ingestChannelConfirm: %s")
		return
	}

	return
}

// LockIngests makes the ingest wait for whatever is happening on the outside, probably creating accounts and such
func (server *OpencxServer) LockIngests() {
	server.ingestMutex.Lock()
}

// UnlockIngests releases the ingest wait for whatever is happening on the outside
func (server *OpencxServer) UnlockIngests() {
	server.ingestMutex.Unlock()
}
