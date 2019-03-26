package cxserver

import (
	"bytes"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/portxo"

	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/lit/btcutil/txscript"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/opencx/logging"
)

// TODO: refactor entire database, match, and asset stuff to support our new automated way of hooks and wallets

// WithdrawCoins inputs the correct parameters to return a withdraw txid
func (server *OpencxServer) WithdrawCoins(address string, pubkey *koblitz.PublicKey, amount uint64, params *coinparam.Params) (txid string, err error) {
	// Create the function, basically make sure the wallet stuff is alright
	var withdrawFunction func(string, *koblitz.PublicKey, uint64) (string, error)
	if withdrawFunction, err = server.withdrawFromChain(params); err != nil {
		err = fmt.Errorf("Error creating withdraw function: \n%s", err)
		return
	}
	// Actually try to withdraw
	if txid, err = withdrawFunction(address, pubkey, amount); err != nil {
		err = fmt.Errorf("Error withdrawing coins: \n%s", err)
		return
	}
	return
}

// withdrawFromChain returns a function that we'll then call from the vtc stuff -- this is a closure that's also a method for server, don't worry about it lol
func (server *OpencxServer) withdrawFromChain(params *coinparam.Params) (withdrawFunction func(string, *koblitz.PublicKey, uint64) (string, error), err error) {

	// Try to get correct wallet
	wallet, found := server.WalletMap[params]
	if !found {
		err = fmt.Errorf("Could not find wallet for those coin params")
		return
	}

	withdrawFunction = func(address string, pubkey *koblitz.PublicKey, amount uint64) (txid string, err error) {

		if amount == 0 {
			err = fmt.Errorf("You can't withdraw 0 %s", params.Name)
			return
		}

		server.LockIngests()
		if err = server.OpencxDB.Withdraw(pubkey, params.Name, amount); err != nil {
			// if errors out, unlock
			server.UnlockIngests()
			return
		}
		server.UnlockIngests()

		// set log level for this thread
		logging.SetLogLevel(2)

		// Decoding given address
		var addr btcutil.Address
		if addr, err = btcutil.DecodeAddress(address, params); err != nil {
			return
		}

		// for paying the other person
		var payToUserScript []byte
		if payToUserScript, err = txscript.PayToAddrScript(addr); err != nil {
			return
		}

		// pick inputs for transaction, idk what the fee shoud be, I think this is the correct byte size but I'm not too sure
		var utxoSlice portxo.TxoSliceByBip69
		var overshoot int64
		if utxoSlice, overshoot, err = wallet.PickUtxos(int64(amount), int64(len(payToUserScript)), 1000, false); err != nil {
			return
		}

		// for giving back the wallet change
		var changeOut *wire.TxOut
		if changeOut, err = wallet.NewChangeOut(overshoot); err != nil {
			return
		}

		// create paytouser txout, we already have change txout from newchangeout
		payToUserTxOut := wire.NewTxOut(int64(amount), payToUserScript)

		// build the transaction
		var withdrawTx *wire.MsgTx
		if withdrawTx, err = wallet.BuildAndSign(utxoSlice, []*wire.TxOut{changeOut, payToUserTxOut}, 0); err != nil {
			return
		}

		buf := new(bytes.Buffer)
		if err = withdrawTx.Serialize(buf); err != nil {
			return
		}

		// Copying and pasting this into the node and sending works
		// The issue where the nodes weren't really adding the tx to the mempool was weird
		// logging.Infof("Serialized transaction: %x\n", buf.Bytes())

		// send out the transaction
		if err = wallet.NewOutgoingTx(withdrawTx); err != nil {
			return
		}

		return withdrawTx.TxHash().String(), nil
	}
	return
}
