package cxserver

import (
	"bytes"
	"fmt"

	"github.com/mit-dci/lit/wallit"

	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/lit/btcutil/txscript"

	"github.com/mit-dci/lit/btcutil/chaincfg"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/opencx/logging"
)

// I'm having fun with closures here to remove all the copy paste that I usually have to do

// VTCWithdraw will be used to watch for events on the chain.
func (server *OpencxServer) VTCWithdraw(address string, username string, amount uint64) (string, error) {
	// Plug in all the specifics -- which wallet to use to send out tx, which testnet, which asset string because I haven't gotten it dependent on params yet
	return server.withdrawFromChain(server.OpencxVTCWallet, &chaincfg.VertcoinTestNetParams, "vtc")(address, username, amount)
}

// BTCWithdraw will be used to watch for events on the chain.
func (server *OpencxServer) BTCWithdraw(address string, username string, amount uint64) (string, error) {
	// Plug in all the specifics -- which wallet to use to send out tx, which testnet, which asset string because I haven't gotten it dependent on params yet
	return server.withdrawFromChain(server.OpencxBTCWallet, &chaincfg.TestNet3Params, "btc")(address, username, amount)
}

// LTCWithdraw will be used to watch for events on the chain.
func (server *OpencxServer) LTCWithdraw(address string, username string, amount uint64) (string, error) {
	// Plug in all the specifics -- which wallet to use to send out tx, which testnet, which asset string because I haven't gotten it dependent on params yet
	return server.withdrawFromChain(server.OpencxLTCWallet, &chaincfg.LiteCoinTestNet4Params, "ltc")(address, username, amount)
}

// withdrawFromChain returns a function that we'll then call from the vtc stuff -- this is a closure that's also a method for server, don't worry about it lol
func (server *OpencxServer) withdrawFromChain(wallet *wallit.Wallit, params *chaincfg.Params, assetString string) func(string, string, uint64) (string, error) {
	var err error
	return func(address string, username string, amount uint64) (string, error) {
		if amount == 0 {
			return "", fmt.Errorf("You can't withdraw 0 %s", assetString)
		}
		var bal uint64

		server.LockIngests()
		if bal, err = server.OpencxDB.GetBalance(username, assetString); err != nil {
			return "", err
		}
		server.UnlockIngests()

		if bal < amount {
			return "", fmt.Errorf("You do not have enough balance to withdraw this amount")
		}

		// set log level for this thread
		logging.SetLogLevel(2)

		// wrap because it needs to be a chaincfg param -- if you use the net name instead of 'btc', 'ltc', 'vtc' then you can stop doing a lot of dumb stuff
		params.PubKeyHashAddrID = wallet.Param.PubKeyHashAddrID

		// Decoding given address
		vtcAddress, err := btcutil.DecodeAddress(address, params)
		if err != nil {
			return "", err
		}

		// for paying the other person
		payToUserScript, err := txscript.PayToAddrScript(vtcAddress)
		if err != nil {
			return "", err
		}

		// pick inputs for transaction, idk what the fee shoud be, I think this is the correct byte size but I'm not too sure
		utxoSlice, overshoot, err := wallet.PickUtxos(int64(amount), int64(len(payToUserScript)), 1000, false)
		if err != nil {
			return "", err
		}

		// for giving back the wallet change
		changeOut, err := wallet.NewChangeOut(overshoot)
		if err != nil {
			return "", err
		}

		// create paytouser txout, we already have change txout from newchangeout
		payToUserTxOut := wire.NewTxOut(int64(amount), payToUserScript)

		// build the transaction
		withdrawTx, err := wallet.BuildAndSign(utxoSlice, []*wire.TxOut{changeOut, payToUserTxOut}, 0)
		if err != nil {
			return "", err
		}

		buf := new(bytes.Buffer)
		if err = withdrawTx.Serialize(buf); err != nil {
			return "", err
		}

		// Copying and pasting this into the node and sending works
		// The issue where the nodes weren't really adding the tx to the mempool was weird
		// logging.Infof("Serialized transaction: %x\n", buf.Bytes())

		// send out the transaction
		if err = wallet.NewOutgoingTx(withdrawTx); err != nil {
			return "", err
		}

		return withdrawTx.TxHash().String(), nil
	}
}
