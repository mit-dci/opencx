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
	server.LockIngests()
	if addressesWeOwn, err = currDepositStore.GetDepositAddressMap(); err != nil {
		// if errors out, unlock
		err = fmt.Errorf("Error getting deposit address map: %s", err)
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}
	server.UnlockIngests()

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
					server.dbLock.Unlock()
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

	// TODO: remove the LockIngests method
	server.LockIngests()
	// So we get the deposits from the deposit store, then we update the settlement engine, then we
	// update the settlement store
	var depositExecs []*match.SettlementExecution
	if depositExecs, err = currDepositStore.UpdateDeposits(deposits, height); err != nil {
		// if errors out, unlock
		err = fmt.Errorf("Error updating deposits while ingesting txs: %s", err)
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}

	var settlementResults []*match.SettlementResult
	for _, setExec := range depositExecs {
		// We always check validity first
		var valid bool
		if valid, err = currSettleEngine.CheckValid(setExec); err != nil {
			err = fmt.Errorf("Error checking exec validity for ingesttx: %s", err)
			server.UnlockIngests()
			server.dbLock.Unlock()
			return
		}

		if valid {
			var setRes *match.SettlementResult
			if setRes, err = currSettleEngine.ApplySettlementExecution(setExec); err != nil {
				err = fmt.Errorf("Error applying settlement exec for ingesttx: %s", err)
				server.UnlockIngests()
				server.dbLock.Unlock()
				return
			}
			settlementResults = append(settlementResults, setRes)
		} else {
			err = fmt.Errorf("Error, invalid settlement exec")
			server.UnlockIngests()
			server.dbLock.Unlock()
			return
		}
	}

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for ingest tx / deposit execs: %s", err)
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}
	server.UnlockIngests()
	server.dbLock.Unlock()

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

	var assetToDebit match.Asset
	if assetToDebit, err = match.AssetFromCoinParam(param); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for ingest channel push: %s", err)
		return
	}

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

	logging.Infof("Confirmed channel from pubkey %x\n", pubkey.SerializeCompressed())
	server.LockIngests()

	setExecForPush := &match.SettlementExecution{
		Type:   match.Debit,
		Asset:  assetToDebit,
		Amount: pushAmt,
	}
	copy(setExecForPush.Pubkey[:], pubkey.SerializeCompressed())

	var setRes *match.SettlementResult
	var valid bool
	if valid, err = currSettleEngine.CheckValid(setExecForPush); err != nil {
		err = fmt.Errorf("Error checking valid exec for ingesting channel push: %s", err)
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}

	var settlementResults []*match.SettlementResult
	if valid {
		var setRes *match.SettlementResult
		if setRes, err = currSettleEngine.ApplySettlementExecution(setExecForPush); err != nil {
			err = fmt.Errorf("Error applying settlement exec for ingesttx: %s", err)
			server.UnlockIngests()
			server.dbLock.Unlock()
			return
		}
		settlementResults = append(settlementResults, setRes)
	} else {
		err = fmt.Errorf("Error, invalid settlement exec")
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}

	if err = currSettleStore.UpdateBalances(settlementResults); err != nil {
		err = fmt.Errorf("Error updating balances for ingest tx / deposit execs: %s", err)
		server.UnlockIngests()
		server.dbLock.Unlock()
		return
	}

	server.UnlockIngests()
	server.dbLock.Unlock()

	return
}

// ingestChannelConfirm changes the user's balance to reflect that a confirmation of a channel happened
func (server *OpencxServer) ingestChannelConfirm(state *qln.StatCom, pubkey *koblitz.PublicKey, coinType uint32) (err error) {
	logging.Infof("Confirmed channel from %x to give me %d of cointype %d\n", pubkey.SerializeCompressed(), state.MyAmt, coinType)

	var param *coinparam.Params
	if param, err = util.GetParamFromHDCoinType(coinType); err != nil {
		return
	}

	logging.Infof("Confirmed channel from pubkey %x\n", pubkey.SerializeCompressed())
	server.LockIngests()

	if err = server.OpencxDB.AddToBalance(pubkey, uint64(state.MyAmt), param); err != nil {
		return
	}

	server.UnlockIngests()

	return
}

// ingestChannelFund only registers the user because the fund channel hasn't necessarily been confirmed yet
func (server *OpencxServer) ingestChannelFund(state *qln.StatCom, pubkey *koblitz.PublicKey, coinType uint32, qchanID uint32) (err error) {
	logging.Infof("Pubkey %x funded a channel to give me %d of cointype %d\n", pubkey.SerializeCompressed(), state.MyAmt, coinType)

	var addrMap map[*coinparam.Params]string
	if addrMap, err = server.GetAddressMap(pubkey); err != nil {
		return
	}

	logging.Infof("Registering user with pubkey %x\n", pubkey.SerializeCompressed())
	server.LockIngests()

	if err = server.OpencxDB.RegisterUser(pubkey, addrMap); err != nil {
		server.UnlockIngests()
		return
	}

	server.UnlockIngests()

	if err = server.SetupFundBack(pubkey, coinType, server.defaultCapacity); err != nil {
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
