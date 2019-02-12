package cxserver

import (
	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/lit/btcutil/txscript"

	"github.com/mit-dci/lit/btcutil/chaincfg"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/opencx/logging"
)

// VTCWalletSend will be used to watch for events on the chain.
func (server *OpencxServer) VTCWalletSend(address string, username string, amount uint64) (string, error) {
	// set log level for this thread
	logging.SetLogLevel(2)

	// wrap chainCfg
	chainCfgWrap := &chaincfg.VertcoinTestNetParams
	chainCfgWrap.PubKeyHashAddrID = server.OpencxVTCWallet.Param.PubKeyHashAddrID

	// Decoding given address
	vtcAddress, err := btcutil.DecodeAddress(address, chainCfgWrap)
	if err != nil {
		return "", err
	}

	// for paying the other person
	payToUserScript, err := txscript.PayToAddrScript(vtcAddress)
	if err != nil {
		return "", err
	}

	// pick inputs for transaction, idk what the fee shoud be, I think this is the correct byte size but I'm not too sure
	utxoSlice, overshoot, err := server.OpencxVTCWallet.PickUtxos(int64(amount), int64(len(payToUserScript)), 1000, false)
	if err != nil {
		return "", err
	}

	// for giving back the wallet change
	changeOut, err := server.OpencxVTCWallet.NewChangeOut(overshoot)
	if err != nil {
		return "", err
	}

	// create paytouser txout, we already have change txout from newchangeout
	payToUserTxOut := wire.NewTxOut(int64(amount), payToUserScript)

	// build the transaction
	withdrawTx, err := server.OpencxVTCWallet.BuildAndSign(utxoSlice, []*wire.TxOut{changeOut, payToUserTxOut}, 0)
	if err != nil {
		return "", err
	}

	// send out the transaction
	if err = server.OpencxVTCWallet.NewOutgoingTx(withdrawTx); err != nil {
		return "", err
	}

	if err = server.OpencxVTCWallet.PushTx(withdrawTx); err != nil {
		return "", err
	}

	return withdrawTx.TxHash().String(), nil
}
