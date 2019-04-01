package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/qln"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wire"
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

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error ingesting transaction list: \n%s", err)
			return
		}
	}()
	var addressesWeOwn map[string]*koblitz.PublicKey
	server.LockIngests()
	if addressesWeOwn, err = server.OpencxDB.GetDepositAddressMap(coinType); err != nil {
		// if errors out, unlock
		server.UnlockIngests()
		logging.Errorf("Error getting deposit address map")
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
					logging.Errorf("Error decoding p2pkh")
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

	server.LockIngests()
	if err = server.OpencxDB.UpdateDeposits(deposits, height, coinType); err != nil {
		// if errors out, unlock
		server.UnlockIngests()
		logging.Errorf("Error updating deposits")
		return
	}
	server.UnlockIngests()

	logging.Debugf("Finished ingesting %s block at height %d", coinType.Name, height)
	if height%10000 == 0 {
		logging.Infof("Finished ingesting %s block at height %d\n", coinType.Name, height)
	}
	return
}

// ingestChannelPush changes the user's balance to reflect that a confirmation of a channel happened
func (server *OpencxServer) ingestChannelPush(pushAmt uint64, pubkey *koblitz.PublicKey, coinType uint32) (err error) {
	logging.Infof("Confirmed push from %x to give me %d of cointype %d\n", pubkey.SerializeCompressed(), pushAmt, coinType)
	var param *coinparam.Params
	if param, err = util.GetParamFromHDCoinType(coinType); err != nil {
		return
	}

	logging.Infof("Confirmed channel from pubkey %x\n", pubkey.SerializeCompressed())
	server.LockIngests()

	if err = server.OpencxDB.AddToBalance(pubkey, pushAmt, param); err != nil {
		return
	}

	server.UnlockIngests()

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
