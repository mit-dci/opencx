package cxserver

import (
	"bytes"
	"fmt"

	"github.com/mit-dci/lit/consts"
	"github.com/mit-dci/lit/lnp2p"
	"github.com/mit-dci/lit/qln"

	"github.com/mit-dci/lit/lnutil"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/portxo"

	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/lit/btcutil/txscript"

	"github.com/mit-dci/lit/btcutil"
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

// withdrawFromChain returns a function that we'll then call from the vtc stuff -- this is a closure that's also a method for server, don't worry about it lol
func (server *OpencxServer) withdrawFromLightning(params *coinparam.Params) (withdrawFunction func(*koblitz.PublicKey, int64) (string, error), err error) {

	// Try to get correct wallet
	// wallet, found := server.WalletMap[params]
	// if !found {
	// 	err = fmt.Errorf("Could not find wallet for those coin params")
	// 	return
	// }

	withdrawFunction = func(pubkey *koblitz.PublicKey, amount int64) (txid string, err error) {

		if amount <= 0 {
			err = fmt.Errorf("Can't withdraw <= 0")
		}

		// calculate fee, do this using subwallet because the funding will all be done through lit
		// TODO: figure out if there is redundancy with server.WalletMap and server.ExchangeNode.SubWallet and
		// if that redundancy is necessary. It might be
		fee := server.ExchangeNode.SubWallet[params.HDCoinType].Fee() * 1000
		if amount < consts.MinOutput+fee {
			err = fmt.Errorf("You can't withdraw any less than %d %s", consts.MinOutput+fee, params.Name)
			return
		}

		var peerIdx uint32
		if peerIdx, err = server.GetPeerFromPubkey(pubkey); err != nil {
			return
		}

		// if we already have a channel and we can, we should push
		if !server.ExchangeNode.ConnectedToPeer(peerIdx) {
			err = fmt.Errorf("Not connected to peer! Please connect to the exchange. We don't know how to connect to you")
			return
		}

		// calculate capacity as a function of the amount to be sent
		var ccap int64
		if amount < consts.MinChanCapacity {
			ccap = consts.MinChanCapacity
		} else {
			ccap = amount + consts.MinOutput + fee
		}

		// TODO: this should only happen when we get a proof that the other person actually took the withdraw / updated the state. We don't have a guarantee that they will always accept

		server.LockIngests()
		if err = server.OpencxDB.Withdraw(pubkey, params.Name, uint64(amount)); err != nil {
			// if errors out, unlock
			server.UnlockIngests()
			return
		}
		server.UnlockIngests()

		// check if any of the channels are of the correct param and have enough capacity (-[min+fee])

		// make data but we don't really want any
		noData := new([32]byte)

		// retreive chanIdx because we need it for qchan for outpoint hash, if that's not useful anymore just make this chanIdx => _
		var chanIdx uint32
		if chanIdx, err = server.ExchangeNode.FundChannel(peerIdx, params.HDCoinType, ccap, amount, *noData); err != nil {
			return
		}

		// get qchan so we can get the outpoint hash
		var qchan *qln.Qchan
		if qchan, err = server.ExchangeNode.GetQchanByIdx(chanIdx); err != nil {
			return
		}

		// get outpoint hash because that's useful information to return
		txid = qchan.Op.Hash.String()

		return
	}
	return
}

// GetPeerFromPubkey gets a peer index from a pubkey.
func (server *OpencxServer) GetPeerFromPubkey(pubkey *koblitz.PublicKey) (peerIdx uint32, err error) {

	var pubkey33 [33]byte
	copy(pubkey33[:], pubkey.SerializeCompressed())
	litAddr := lnutil.LitAdrFromPubkey(pubkey33)

	// until this is removed, this is good for our purposes
	if peerIdx, err = server.ExchangeNode.FindPeerIndexByAddress(litAddr); err != nil {
		return
	}

	return
}

// GetQchansByPeerParam gets channel indexes for a param and pubkey / peer
func (server *OpencxServer) GetQchansByPeerParam(pubkey *koblitz.PublicKey, param *coinparam.Params) (qchans []*qln.Qchan, err error) {

	// get the peer idx
	var peerIdx uint32
	if peerIdx, err = server.GetPeerFromPubkey(pubkey); err != nil {
		return
	}

	// get the peer
	var peer *lnp2p.Peer
	if peer = server.ExchangeNode.PeerMan.GetPeerByIdx(int32(peerIdx)); err != nil {
		return
	}

	// lock this so we can be in peace
	server.ExchangeNode.PeerMapMtx.Lock()
	// get the remote peer from the qchan
	var remotePeer *qln.RemotePeer
	var found bool
	if remotePeer, found = server.ExchangeNode.PeerMap[peer]; !found {
		err = fmt.Errorf("Could not find remote peer that peer manager is tracking, there's something wrong with the lit node")
		// unlock because we have to before we return or else we deadlock
		server.ExchangeNode.PeerMapMtx.Unlock()
		return
	}
	server.ExchangeNode.PeerMapMtx.Unlock()

	// populate qchans just to be safe -- this might not be necessary and makes this function very inefficient
	if err = server.ExchangeNode.PopulateQchanMap(remotePeer); err != nil {
		return
	}

	// get qchans from peer
	for _, qchan := range remotePeer.QCs {
		// if this is the same coin then return the idx
		if qchan.Coin() == param.HDCoinType {
			qchans = append(qchans, qchan)
		}
	}

	return
}
