package cxserver

import (
	"fmt"
	"os"
	"time"

	"github.com/mit-dci/lit/portxo"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/consts"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/litrpc"
	"github.com/mit-dci/lit/qln"
	"github.com/mit-dci/opencx/logging"
)

// SetupLitNode sets up the lit node for use later, I wanna do this because this really shouldn't be in initialization code? should it be?
// basically just run this after you unlock the key
func (server *OpencxServer) SetupLitNode(privkey *[32]byte, subDirName string, trackerURL string, proxyURL string, nat string) (err error) {

	// create lit root directory
	if _, err = os.Stat(server.OpencxRoot + subDirName); os.IsNotExist(err) {
		if err = os.Mkdir(server.OpencxRoot+subDirName, 0700); err != nil {
			logging.Errorf("Error while creating a directory: \n%s", err)
		}
	}

	if server.ExchangeNode, err = qln.NewLitNode(privkey, server.OpencxRoot+subDirName, trackerURL, proxyURL, nat); err != nil {
		return
	}

	return
}

// SetupLitRPCConnect sets up an rpc connection with a running lit node?
func (server *OpencxServer) SetupLitRPCConnect(rpchost string, rpcport uint16) {
	var err error
	defer func() {
		if err != nil {
			logging.Errorf("Error creating lit RPC connection: \n%s", err)
		}
	}()
	if server.ExchangeNode == nil {
		err = fmt.Errorf("Please start the exchange node before trying to create its RPC")
		return
	}

	rpc1 := new(litrpc.LitRPC)
	rpc1.Node = server.ExchangeNode
	rpc1.OffButton = make(chan bool, 1)
	server.ExchangeNode.RPC = rpc1

	// we don't care about unauthRPC
	go litrpc.RPCListen(rpc1, rpchost, rpcport)

	<-rpc1.OffButton
	logging.Infof("Got stop request\n")
	time.Sleep(time.Second)
	return
}

// SetupFundBack funds a node back after a sigproof
func (server *OpencxServer) SetupFundBack(pubkey *koblitz.PublicKey, currCoinType uint32, channelCapacity int64) (err error) {

	for _, param := range server.CoinList {
		if param.HDCoinType != currCoinType {
			var pWallet qln.UWallet
			var found bool
			if pWallet, found = server.ExchangeNode.SubWallet[param.HDCoinType]; !found {
				err = fmt.Errorf("Don't have wallet for coin, how did this happen?")
				return
			}

			var allUtxos []*portxo.PorTxo
			if allUtxos, err = pWallet.UtxoDump(); err != nil {
				return
			}

			var totalValue int64
			for _, utxo := range allUtxos {
				totalValue += utxo.Value
			}
			logging.Infof("Total value in coin %s: %d", param.Name, totalValue)
			var txid string
			// Send with 0 balance on their side
			logging.Debugf("Creating channel with %x for %d %s coins.", pubkey.SerializeCompressed(), channelCapacity, param.Name)
			if txid, err = server.CreateChannel(pubkey, 0, channelCapacity, param); err != nil {
				logging.Errorf("Creating %s channel not successful", param.Name)
				return
			}
			logging.Debugf("Outpoint hash for fund back channel: %s\n", txid)
		}
	}

	return
}

// CreateChannel creates a channel with pubkey and will send a certain amount on creation.
// if the send amount is not 0 then it will withdraw from the
func (server *OpencxServer) CreateChannel(pubkey *koblitz.PublicKey, initSend int64, ccap int64, params *coinparam.Params) (txid string, err error) {
	if initSend < 0 {
		err = fmt.Errorf("Can't withdraw <= 0")
		return
	}

	// calculate fee, do this using subwallet because the funding will all be done through lit
	// TODO: figure out if there is redundancy with server.WalletMap and server.ExchangeNode.SubWallet and
	// if that redundancy is necessary. It might be
	fee := server.ExchangeNode.SubWallet[params.HDCoinType].Fee() * 1000
	if initSend != 0 && initSend < consts.MinOutput+fee {
		err = fmt.Errorf("You can't withdraw any less than %d %s", consts.MinOutput+fee, params.Name)
		return
	}

	var peerIdx uint32
	if peerIdx, err = server.GetPeerFromPubkey(pubkey); err != nil {
		err = fmt.Errorf("You may not have ever connected with the exchange, or you're using a different identity. The exchange can only authenticate for channel creating if you are the node: \n%s", err)
		return
	}

	logging.Debugf("Checking if connected to peer")

	// if we already have a channel and we can, we should push
	if !server.ExchangeNode.ConnectedToPeer(peerIdx) {
		err = fmt.Errorf("Not connected to peer! Please connect to the exchange. We don't know how to connect to you")
		return
	}

	// TODO: this should only happen when we get a proof that the other person actually took the withdraw / updated the state. We don't have a guarantee that they will always accept

	if initSend != 0 {
		logging.Debugf("Checking withdraw lock...")
		server.LockIngests()
		logging.Debugf("Locked ingests, withdrawing")
		if err = server.OpencxDB.Withdraw(pubkey, params.Name, uint64(initSend)); err != nil {
			// if errors out, unlock
			logging.Errorf("Error while withdrawing from db: %s\n", err)
			server.UnlockIngests()
			return
		}
		server.UnlockIngests()
	}

	var utxoDump []*portxo.PorTxo
	if utxoDump, err = server.ExchangeNode.SubWallet[params.HDCoinType].UtxoDump(); err != nil {
		logging.Infof("Error dumping utxos")
		return
	}

	for _, utxo := range utxoDump {
		logging.Infof("Coin %s UTXO value: %d\n", params.Name, utxo.Value)
	}

	// check if any of the channels are of the correct param and have enough capacity (-[min+fee])

	// make data but we don't really want any
	noData := new([32]byte)

	logging.Debugf("Trying to fund channel")
	// retreive chanIdx because we need it for qchan for outpoint hash, if that's not useful anymore just make this chanIdx => _
	var chanIdx uint32
	if chanIdx, err = server.ExchangeNode.FundChannel(peerIdx, params.HDCoinType, ccap, initSend, *noData); err != nil {
		return
	}

	logging.Debugf("Getting qchanidx")
	// get qchan so we can get the outpoint hash
	var qchan *qln.Qchan
	if qchan, err = server.ExchangeNode.GetQchanByIdx(chanIdx); err != nil {
		return
	}

	logging.Debugf("We're pretty much done with this withdraw")
	// get outpoint hash because that's useful information to return
	txid = qchan.Op.Hash.String()

	return
}
