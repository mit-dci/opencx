package cxserver

import (
	"fmt"
	"math"

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
	// addressesWeOwn, err := server.OpencxDB.GetDepositAddressMap(coinType)
	// if err != nil {
	// 	return
	// }

	var deposits []match.Deposit

	for _, tx := range txList {
		for _, output := range tx.TxOut {

			scriptType, data := util.ScriptType(output.PkScript)
			if scriptType == "P2PKH" {
				var addr string
				if addr, err = util.NewAddressPubKeyHash(data, coinType); err != nil {
					return
				}

				if name, depErr := server.OpencxDB.GetDepositName(addr, coinType); depErr == nil && name != "" {
					newDeposit := match.Deposit{
						Name:                name,
						Address:             addr,
						Amount:              uint64(output.Value),
						Txid:                tx.TxHash().String(),
						CoinType:            coinType,
						BlockHeightReceived: height,
						Confirmations:       6 * uint64(math.Pow(6, float64(output.Value)/math.Pow10(8))),
					}

					logging.Infof("%s\n", newDeposit.String())
					deposits = append(deposits, newDeposit)
				} else if depErr != nil {
					err = depErr
					return
				}
			}
		}
	}

	if err = server.OpencxDB.UpdateDeposits(deposits, height, coinType); err != nil {
		return
	}

	if height%100 == 0 {
		logging.Infof("Finished ingesting %s block at height %d. There were %d deposits.\n", coinType.Name, height, len(deposits))
	}
	return nil
}
