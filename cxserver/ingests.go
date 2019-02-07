package cxserver

import (
	"fmt"

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
	server.ingestMutex.Lock()
	addressesWeOwn, err := server.OpencxDB.GetDepositAddressMap(coinType)
	if err != nil {
		return
	}

	var deposits []match.Deposit

	for _, tx := range txList {
		for _, output := range tx.TxOut {

			scriptType, data := util.ScriptType(output.PkScript)
			if scriptType == "P2PKH" {
				var addr string
				if addr, err = util.NewAddressPubKeyHash(data, coinType); err != nil {
					return
				}

				// if coinType.Name == "vtcreg" {
				// 	logging.Infof("Generated address: %s\n", addr)
				// 	logging.Infof("Address thing we own; %s\n", addressesWeOwn)
				// }

				if name, found := addressesWeOwn[addr]; found {
					newDeposit := match.Deposit{
						Name:                name,
						Address:             addr,
						Amount:              uint64(output.Value),
						Txid:                tx.TxHash().String(),
						CoinType:            coinType,
						BlockHeightReceived: height,
						Confirmations:       6,
					}

					logging.Infof("%s\n", newDeposit.String())
					deposits = append(deposits, newDeposit)
				}
			}
		}
	}

	if err = server.OpencxDB.UpdateDeposits(deposits, height, coinType); err != nil {
		return
	}

	if height%100 == 0 {
		logging.Infof("Finished ingesting %s block at height %d\n", coinType.Name, height)
	}
	server.ingestMutex.Unlock()
	return nil
}

// LockIngests makes the ingest wait for whatever is happening on the outside, probably creating accounts and such
func (server *OpencxServer) LockIngests() {
	server.ingestMutex.Lock()
}

// UnlockIngests releases the ingest wait for whatever is happening on the outside
func (server *OpencxServer) UnlockIngests() {
	server.ingestMutex.Unlock()
}
