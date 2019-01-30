package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/bech32"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wire"
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

			// TODO: need to write some stuff to check what kind of script it is, because the current stuff doesn't have enough, or I'm just very wrong
			addFromPkH, err := bech32.SegWitAddressEncode(coinType.Bech32Prefix, output.PkScript)
			if err != nil {
				fmt.Printf("Height: %d, Non P2PKH/P2WPKH/P2WSH output detected, txid: %s\n", height, tx.TxHash())
			} else {
				fmt.Printf("Address %s received %d testnet satoshis in txid %s\n", addFromPkH, output.Value, tx.TxHash())
			}
			fmt.Printf("pkScript: %x\n", output.PkScript)

			// fmt.Printf("Address: %s\n", txscript.ExtractPkScriptAddrs(output.PkScript, coinType))
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
