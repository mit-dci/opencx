package cxserver

import (
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/qln"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/opencx/logging"
)

// GetPushHandler gets the handler func to add to the user's balance since they've pushed over a lightning channel
func (server *OpencxServer) GetPushHandler() (hFunc func(event eventbus.Event) eventbus.EventHandleResult) {
	hFunc = func(event eventbus.Event) (res eventbus.EventHandleResult) {
		// We know this is a channel state update event
		ee, ok := event.(qln.ChannelStateUpdateEvent)
		if !ok {
			logging.Errorf("Wrong type of event, why are you making this the handler for that?")
			// Still don't know if this is the right thing to return when we have an error
			return eventbus.EHANDLE_CANCEL
		}

		logging.Infof("Their pubkey: %x", (&ee.TheirPub).SerializeCompressed())

		var err error
		if !ee.State.Failed {
			if err = server.ingestChannelPush(uint64(ee.State.Delta), &ee.TheirPub, ee.CoinType); err != nil {
				logging.Errorf("ingesting channel push error: %s", err)
				return eventbus.EHANDLE_CANCEL
			}
		}

		return eventbus.EHANDLE_OK
	}

	return
}

// GetOPConfirmHandler gets the handler func to pass in an amount to the updatebalance function
func (server *OpencxServer) GetOPConfirmHandler() (hFunc func(event eventbus.Event) eventbus.EventHandleResult) {
	hFunc = func(event eventbus.Event) (res eventbus.EventHandleResult) {
		// We know this is a channel state update event
		ee, ok := event.(qln.ChannelStateUpdateEvent)
		if !ok {
			logging.Errorf("Wrong type of event, why are you making this the handler for that?")
			// Still don't know if this is the right thing to return when we have an error
			return eventbus.EHANDLE_CANCEL
		}

		var err error
		if !ee.State.Failed {
			if err = server.ingestChannelConfirm(ee.State, &ee.TheirPub, ee.CoinType); err != nil {
				logging.Errorf("ingesting channel confirm error: %s", err)
				return eventbus.EHANDLE_CANCEL
			}
		}

		return eventbus.EHANDLE_OK
	}
	return
}

// GetSigProofHandler gets the handler func to pass in things to the register function. We get their pub key from the
// QChan through the event bus and then use that to register a new user.
func (server *OpencxServer) GetSigProofHandler() (hFunc func(event eventbus.Event) eventbus.EventHandleResult) {
	hFunc = func(event eventbus.Event) (res eventbus.EventHandleResult) {
		// We know this is a channel state update event
		ee, ok := event.(qln.ChannelStateUpdateEvent)
		if !ok {
			logging.Errorf("Wrong type of event, why are you making this the handler for that?")
			// Still don't know if this is the right thing to return
			return eventbus.EHANDLE_CANCEL
		}

		var err error
		if !ee.State.Failed {
			logging.Infof("Channel not failed, ingest")
			if err = server.ingestChannelFund(ee.State, &ee.TheirPub, ee.CoinType, ee.ChanIdx); err != nil {
				logging.Errorf("ingesting channel fund error: %s", err)
				return eventbus.EHANDLE_CANCEL
			}
		}

		return eventbus.EHANDLE_OK
	}
	return
}

// HeightHandler is a handler for when there is a height and block event for the wallet. We need both channels to work and be synchronized, which I'm assuming is the case in the lit repos. Will need to double check.
func (server *OpencxServer) HeightHandler(incomingBlockHeight chan lnutil.HeightEvent, blockChan chan *wire.MsgBlock, coinType *coinparam.Params) {
	for {
		h := <-incomingBlockHeight
		block := <-blockChan
		server.CallIngest(h.Height, block, coinType)
	}
}

// ChainHookHeightHandler is a handler for when there is a height and block event. We need both channels to work and be synchronized, which I'm assuming is the case in the lit repos. Will need to double check.
func (server *OpencxServer) ChainHookHeightHandler(incomingBlockHeight chan int32, blockChan chan *wire.MsgBlock, coinType *coinparam.Params) {
	for {

		// this used to be commented out. Since in lit the channels are buffered, we HAVE to make sure that this is cleared
		// otherwise lit will just completely block and wait for us to pull from the channel, and we will stop getting
		// headers and everything. IF it's needed, just always pulling from this would be fine if we don't care about it.
		h := <-incomingBlockHeight
		block := <-blockChan
		logging.Infof("Block %s from %s", block.Header.BlockHash(), coinType.Name)
		server.CallIngest(h, block, coinType)
	}
}

// CallIngest calls the ingest function. This is so we can make a bunch of different handlers that call this depending on which way they use channels.
func (server *OpencxServer) CallIngest(blockHeight int32, block *wire.MsgBlock, coinType *coinparam.Params) {
	// logging.Debugf("Ingesting %d transactions at height %d\n", len(block.Transactions), blockHeight)
	if err := server.ingestTransactionListAndHeight(block.Transactions, uint64(blockHeight), coinType); err != nil {
		logging.Infof("something went horribly wrong with %s\n", coinType.Name)
		logging.Errorf("Here's what went horribly wrong: %s\n", err)
	}
}
