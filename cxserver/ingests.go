package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/btcutil/chaincfg"

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

				coinTypeChaincfgWrap := new(chaincfg.Params)
				coinTypeChaincfgWrap.PubKeyHashAddrID = coinType.PubKeyHashAddrID
				if addr, err = btcutil.NewAddressPubKeyHash(data, coinTypeChaincfgWrap); err != nil {
					return err
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
	return nil
}

func (server *OpencxServer) ingestChannelFund() (err error) {
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
