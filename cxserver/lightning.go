package cxserver

import (
	"crypto/rand"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/btcsuite/fastsha256"
	"github.com/Rjected/lit/lncore"
	"github.com/Rjected/lit/lnutil"
	"github.com/Rjected/lit/portxo"

	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/consts"
	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/Rjected/lit/litrpc"
	"github.com/Rjected/lit/qln"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
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

// CreateSwap creates a swap with the user depending on an order specified. This is the main functionality for non custodial exchange.
// TODO: check over this code, it's a large function
func (server *OpencxServer) CreateSwap(pubkey *koblitz.PublicKey, order *match.LimitOrder) (err error) {
	// TODO

	// get all the channels
	var channels []*qln.Qchan
	if channels, err = server.ExchangeNode.GetAllQchans(); err != nil {
		err = fmt.Errorf("Error getting channels for swap: %s", err)
		return
	}

	// We need to use this address to get a bunch of stuff from the lit node
	var pubkey33 [33]byte

	copy(pubkey33[:], pubkey.SerializeCompressed())
	lnAddrString := lnutil.LitAdrFromPubkey(pubkey33)

	var address lncore.LnAddr
	if address, err = lncore.ParseLnAddr(lnAddrString); err != nil {
		err = fmt.Errorf("Error parsing ln address for swap: %s", err)
		return
	}

	// TODO: do something else to manage peer stuff. We need an identity key to be
	// associated with a channel somehow, but right now we're just using peers

	// Get this pubkey's peer index.
	thisPeer := server.ExchangeNode.PeerMan.GetPeer(address).GetIdx()

	// figure out which ones match the coins for this order that we care about
	var assetWantChannels []*qln.Qchan
	var assetWantOurCap uint64
	var assetHaveChannels []*qln.Qchan
	var assetHaveTheirCap uint64

	// get the coin parameters for the trading pair
	var wantAsset *coinparam.Params
	if wantAsset, err = order.TradingPair.AssetWant.CoinParamFromAsset(); err != nil {
		err = fmt.Errorf("Error getting assetwant coin from order for swap: %s", err)
		return
	}

	var haveAsset *coinparam.Params
	if haveAsset, err = order.TradingPair.AssetHave.CoinParamFromAsset(); err != nil {
		err = fmt.Errorf("Error getting assethave coin from order for swap: %s", err)
		return
	}

	// TODO: This code is the best we have right now as far as optimizing channel
	// capacity goes. If looking for a place to start, this is it.

	// go through all of the channels, skip the ones that aren't associated with this peer
	for _, channel := range channels {

		// This is what we're doing to say "does the pubkey match?", the
		// pubkey that gives commands does not have to be the pubkey for the channel. While
		// this would probably be ideal if channels already existed, they don't always exist
		// and sometimes there are multiple channels, each with different channel pubkeys,
		// that belong to the same user that we need to use. So for this the identity key
		// is what we should be using.
		if channel.Peer() == thisPeer {

			myAmt, theirAmt := channel.GetChannelBalances()

			if channel.Coin() == wantAsset.HDCoinType {
				// count only the coins we would be able to send
				if myAmt-consts.MinOutput > 0 {
					assetWantChannels = append(assetWantChannels, channel)
					// we can't send so much as to bring our output below the minoutput
					assetWantOurCap += uint64(myAmt - consts.MinOutput)
				}
			}

			if channel.Coin() == haveAsset.HDCoinType {
				// Count only the coins they would be able to send
				if theirAmt-consts.MinOutput > 0 {
					assetHaveChannels = append(assetHaveChannels, channel)
					// they can't send so much as to bring their output below the minoutput
					assetHaveTheirCap += uint64(theirAmt - consts.MinOutput)
				}
			}
		}
	}

	// Let's sort the channels according to capacity, lowest first. MyAmt for the channels
	// we're going to be pushing from (assetWant), and theirAmt for channels we're
	// receiving from

	// sort assetWantChannels
	sort.Slice(assetWantChannels, func(i, j int) bool {
		myAmti, _ := assetWantChannels[i].GetChannelBalances()
		myAmtj, _ := assetWantChannels[j].GetChannelBalances()
		return myAmti < myAmtj
	})

	// sort assetHaveChannels
	sort.Slice(assetHaveChannels, func(i, j int) bool {
		_, theirAmti := assetHaveChannels[i].GetChannelBalances()
		_, theirAmtj := assetHaveChannels[j].GetChannelBalances()
		return theirAmti < theirAmtj
	})

	// Create an arbitrary locktime lol
	locktime := uint32(100)
	// Create a preimage
	var R [16]byte
	// Read random bytes
	if _, err = rand.Read(R[:]); err != nil {
		err = fmt.Errorf("Error reading random bytes into preimage: %s", err)
		return
	}
	RHash := fastsha256.Sum256(R[:])

	// TODO: Take the below code and abstract it into a method for general spending
	// and withdraw. The channel selection algorithm applies

	// Set up HTLCs from us to them. We assume that the channels can be pushed to
	// since we didn't add the ones that couldn't be pushed from, for either side
	// Use the channels we have, saving the largest for last, and if we have anything left
	// we'll then create a new funding transaction pushing funds with an HTLC set up
	// already.
	amountRemainingWant := order.AmountWant
	for i := 0; amountRemainingWant > 0 && i < len(assetWantChannels); i++ {
		myAmt, _ := assetWantChannels[i].GetChannelBalances()

		// If the amount of this channel is <= the amount remaining, then send the full
		// available amount (minus minoutput)
		var avail uint32

		// Cast a bunch, TODO make sure all this casting is safe
		if uint64(myAmt-consts.MinOutput) < amountRemainingWant {
			// send as much as we can
			avail = uint32(myAmt - consts.MinOutput)
		} else {
			avail = uint32(amountRemainingWant)
		}
		// We don't have any data to send
		if err = server.ExchangeNode.OfferHTLC(assetWantChannels[i], avail, RHash, locktime, [32]byte{}); err != nil {
			err = fmt.Errorf("Error offering HTLC for atomic swap: %s", err)
			return
		}

		// we sent avail amount, subtract from amount remaining
		amountRemainingWant -= uint64(avail)
	}

	// if we still have some left at this point, set up a funding transaction and
	// push just enough in an HTLC for this swap to work.
	if amountRemainingWant > 0 {
		var wallet qln.UWallet
		var ok bool
		if wallet, ok = server.ExchangeNode.SubWallet[wantAsset.HDCoinType]; !ok {
			err = fmt.Errorf("No wallet of the %d type connected", wantAsset.HDCoinType)
			return
		}
		fee := wallet.Fee() * 1000
		// Set the channel capacity to the amount remaining + fee if they're greater than
		// the min capacity, otherwise set the capacity to the min capacity

		// we add the min output because when we send, we'll need some left on our side.
		var newChanIdx uint32
		desiredChannelCapacity := consts.MinOutput + amountRemainingWant + uint64(fee)
		if desiredChannelCapacity < uint64(consts.MinChanCapacity) {
			desiredChannelCapacity = uint64(consts.MinChanCapacity)
		}

		// no data
		if newChanIdx, err = server.ExchangeNode.FundChannel(thisPeer, wantAsset.HDCoinType, consts.MinChanCapacity, 0, [32]byte{}); err != nil {
			err = fmt.Errorf("Could not fund channel for final send: %s", err)
			return
		}

		var newlyFundedChannel *qln.Qchan
		if newlyFundedChannel, err = server.ExchangeNode.GetQchanByIdx(newChanIdx); err != nil {
			err = fmt.Errorf("Error getting newly funded channel by index for swap: %s", err)
			return
		}

		// send amountremainingwant in htlc
		// We don't have any data to send
		if err = server.ExchangeNode.OfferHTLC(newlyFundedChannel, uint32(amountRemainingWant), RHash, locktime, [32]byte{}); err != nil {
			err = fmt.Errorf("Error offering final send HTLC for atomic swap: %s", err)
			return
		}
	}

	// Ask for HTLC's from them to us
	amountRemainingHave := order.AmountHave
	for i := 0; amountRemainingHave > 0 && i < len(assetHaveChannels); i++ {
		_, theirAmt := assetHaveChannels[i].GetChannelBalances()

		// If the amount of this channel is <= the amount remaining, then send the full
		// available amount (minus minoutput)
		var avail uint32

		// Cast a bunch, TODO make sure all this casting is safe
		if uint64(theirAmt-consts.MinOutput) < amountRemainingHave {
			// send as much as we can
			avail = uint32(theirAmt - consts.MinOutput)
		} else {
			avail = uint32(amountRemainingHave)
		}

		// TODO: we need a requestHTLC method, sort of like the other side of dual fund
		// We don't have any data to send
		// if err = server.ExchangeNode.OfferHTLC(assetHaveChannels[i], avail, RHash, locktime, [32]byte{}); err != nil {
		// 	err = fmt.Errorf("Error offering HTLC for atomic swap: %s", err)
		// 	return
		// }

		// we sent avail amount, subtract from amount remaining
		amountRemainingHave -= uint64(avail)
	}

	// if we still have some left at this point, set up a funding transaction and
	// push just enough in an HTLC for this swap to work.
	if amountRemainingHave > 0 {
		var wallet qln.UWallet
		var ok bool
		if wallet, ok = server.ExchangeNode.SubWallet[haveAsset.HDCoinType]; !ok {
			err = fmt.Errorf("No wallet of the %d type connected", haveAsset.HDCoinType)
			return
		}
		fee := wallet.Fee() * 1000
		// Set the channel capacity to the amount remaining + fee if they're greater than
		// the min capacity, otherwise set the capacity to the min capacity

		// we add the min output because when we send, we'll need some left on our side.
		var newChanResult *qln.DualFundingResult
		desiredChannelCapacity := consts.MinOutput + amountRemainingHave + uint64(fee)
		if desiredChannelCapacity < uint64(consts.MinChanCapacity) {
			desiredChannelCapacity = uint64(consts.MinChanCapacity)
		}

		// no data
		if newChanResult, err = server.ExchangeNode.DualFundChannel(thisPeer, haveAsset.HDCoinType, 0, consts.MinChanCapacity); err != nil {
			err = fmt.Errorf("Could not fund channel w/ reason %d for final send: %s", newChanResult.DeclineReason, err)
			return
		}

		if newChanResult.Error {
			err = fmt.Errorf("Error creating a dual funding channel. reason: %d. accepted: %t", newChanResult.DeclineReason, newChanResult.Accepted)
			return
		}

		// TODO: Use PayMultihop - Make sure exchange rate stays the same
		// var newlyFundedChannel *qln.Qchan
		// if newlyFundedChannel, err = server.ExchangeNode.GetQchanByIdx(newChanResult.ChannelId); err != nil {
		// 	err = fmt.Errorf("Error getting newly funded channel by index for swap: %s", err)
		// 	return
		// }

		// TODO: Again, use PayMultihop
		// send amountremaininghave in htlc
		// We don't have any data to send
		// if err = server.ExchangeNode.OfferHTLC(newlyFundedChannel, uint32(amountRemainingHave), RHash, locktime, [32]byte{}); err != nil {
		// 	err = fmt.Errorf("Error offering final send HTLC for atomic swap: %s", err)
		// 	return
		// }
	}

	return
}

// SetupFundBack funds a node back after a sigproof
func (server *OpencxServer) SetupFundBack(pubkey *koblitz.PublicKey, currCoinType uint32, channelCapacity int64) (err error) {

	// now check for all the settlement layers again... since lightning itself is a settlement layer
	// maybe we should abstract
	for param, _ := range server.SettlementEngines {
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
		if err = server.CreditUser(pubkey, uint64(initSend), params); err != nil {
			err = fmt.Errorf("Error while crediting user for CreateChannel: %s\n", err)
			return
		}
	}

	var thisWallet qln.UWallet
	var ok bool
	if thisWallet, ok = server.ExchangeNode.SubWallet[params.HDCoinType]; !ok {
		err = fmt.Errorf("Cound not find subwallet for CreateChannel")
		return
	}

	var utxoDump []*portxo.PorTxo
	if utxoDump, err = thisWallet.UtxoDump(); err != nil {
		logging.Infof("Error dumping utxos")
		err = fmt.Errorf("Error dumping utxos for subwallet for CreateChannel: %s", err)
		return
	}

	for _, utxo := range utxoDump {
		logging.Infof("Coin %s UTXO value: %d\n", params.Name, utxo.Value)
	}

	// check if any of the channels are of the correct param and have enough capacity (-[min+fee])

	// make data but we don't really want any
	noData := new([32]byte)

	logging.Debugf("Trying to fund channel")
	// retrieve chanIdx because we need it for qchan for outpoint hash, if that's not useful anymore just make this chanIdx => _
	var chanIdx uint32
	if chanIdx, err = server.ExchangeNode.FundChannel(peerIdx, params.HDCoinType, ccap, initSend, *noData); err != nil {
		err = fmt.Errorf("Error funding channel for CreateChannel: %s", err)
		return
	}

	logging.Debugf("Getting qchanidx")
	// get qchan so we can get the outpoint hash
	var qchan *qln.Qchan
	if qchan, err = server.ExchangeNode.GetQchanByIdx(chanIdx); err != nil {
		err = fmt.Errorf("Error getting qchan by idx for CreateChannel: %s", err)
		return
	}

	logging.Debugf("We're pretty much done with this withdraw")
	// get outpoint hash because that's useful information to return
	txid = qchan.Op.Hash.String()

	return
}
