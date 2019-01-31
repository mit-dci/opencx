package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/opencx/util"
)

// IngestTransactionListAndHeight processes a transaction list and corresponding height
func (server *OpencxServer) ingestTransactionListAndHeight(txList []*wire.MsgTx, height int32, coinType coinparam.Params) error {
	// get list of addresses we own
	// check the sender, amounts, receiver of all the transactions
	// check if the receiver is us
	// if so, add the deposit to the table, create a # of confirmations past the height at which it was received
	amounts := make([]uint64, len(txList))
	for _, tx := range txList {
		for _, output := range tx.TxOut {

			scriptType, data := util.ScriptType(output.PkScript)
			if scriptType == "P2PKH" {
				// fmt.Printf("Script: %x\n", output.PkScript)
				// fmt.Printf("Data: %x\n", data)
				_, err := util.NewAddressPubKeyHash(data, &coinType)
				if err != nil {
					return fmt.Errorf("Error converting pubkeyhash into address while ingesting transaction: \n%s", err)
				}

				// fmt.Printf("Address %s got %f BTC in tx %s\n", addr, float64(output.Value)/(math.Pow10(8)), tx.TxHash().String())
			} else {
				// fmt.Printf("Script type: %s\n", scriptType)
			}
		}
		amounts = append(amounts, sumTxOut(tx.TxOut))
	}
	err := server.OpencxDB.UpdateDeposits(amounts, coinType)
	if err != nil {
		return fmt.Errorf("Error while ingesting transaction list and height: \n%s", err)
	}
	return nil
}

func sumTxOut(outputs []*wire.TxOut) uint64 {
	var amount uint64
	for _, output := range outputs {
		amount += uint64(output.Value)
	}
	return amount
}
