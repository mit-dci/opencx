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
	server.LockIngests()
	addressesWeOwn, err := server.OpencxDB.GetDepositAddressMap(coinType)
	if err != nil {
		// if errors out, unlock
		server.UnlockIngests()
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
		return
	}
	server.UnlockIngests()

	logging.Debugf("Finished ingesting %s block at height %d", coinType.Name, height)
	if height%10000 == 0 {
		logging.Infof("Finished ingesting %s block at height %d\n", coinType.Name, height)
	}
	return
}

func (server *OpencxServer) ingestChannelFund(state *qln.StatCom, pubkey *koblitz.PublicKey) (err error) {
	logging.Infof("Pubkey %x funded a channel to give me %d\n", pubkey.SerializeCompressed(), state.MyAmt)

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
