package cxserver

import (
	"github.com/mit-dci/lit/btcutil/base58"

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

	// creating derived key from root
	// sha := sha256.New()
	// sha.Write([]byte(username))

	// TODO: Make this better lol, one of the most annoying things (and now that I think about it, exchanges must have this down to a T) is key assignment and management
	// We mod by 0x80000000 to make sure it's not hardened
	// data := binary.BigEndian.Uint32(sha.Sum(nil)[:]) % 0x80000000

	// wrap chainCfg
	chainCfgWrap := &chaincfg.VertcoinTestNetParams
	chainCfgWrap.PubKeyHashAddrID = server.OpencxVTCWallet.Param.PubKeyHashAddrID

	// for debugging
	addrs, err := server.OpencxBTCWallet.AdrDump()
	for _, addr := range addrs {
		logging.Infof("Address: %s\n", base58.CheckEncode(addr[:], server.OpencxVTCWallet.Param.PubKeyHashAddrID))
	}

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

	logging.Infof("Amount: %d", amount)
	logging.Infof("int64 amount: %d", int64(amount))
	// pick inputs for transaction, idk what the fee shoud be, I think this is the correct byte size but I'm not too sure
	utxoSlice, overshoot, err := server.OpencxVTCWallet.PickUtxos(int64(amount), int64(len(payToUserScript)), 1000, false)
	if err != nil {
		return "", err
	}
	logging.Infof("Picked inputs for utxos")

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
	logging.Infof("Built and signed withdraw transaction")
	for _, elem := range withdrawTx.TxOut {
		logging.Infof("Output amount: %d", elem.Value)
	}

	for _, elem := range withdrawTx.TxIn {
		logging.Infof("Input txid: %s", elem.PreviousOutPoint.Hash.String())
	}

	// send out the transaction
	if err = server.OpencxVTCWallet.NewOutgoingTx(withdrawTx); err != nil {
		return "", err
	}

	return withdrawTx.TxHash().String(), nil
}
