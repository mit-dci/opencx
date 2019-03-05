package cxserver

import (
	"fmt"
	"sync"

	"github.com/mit-dci/lit/eventbus"

	"github.com/mit-dci/lit/btcutil"
	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/qln"
	"github.com/mit-dci/lit/wallit"
	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB   cxdb.OpencxStore
	OpencxRoot string
	OpencxPort uint16
	AssetArray []match.Asset
	PairsArray []*match.Pair
	// Hehe it's the vault, pls don't steal
	OpencxBTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxVTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxLTCTestPrivKey *hdkeychain.ExtendedKey

	ExchangeNode *qln.LitNode

	BlockChanMap       map[int]chan *wire.MsgBlock
	HeightEventChanMap map[int]chan lnutil.HeightEvent
	ingestMutex        sync.Mutex

	OpencxBTCWallet *wallit.Wallit
	OpencxLTCWallet *wallit.Wallit
	OpencxVTCWallet *wallit.Wallit

	orderMutex *sync.Mutex
	OrderMap   map[match.Pair][]*match.LimitOrder
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}

// LockOrders locks the order mutex
func (server *OpencxServer) LockOrders() {
	server.orderMutex.Lock()
}

// UnlockOrders unlocks the order mutex
func (server *OpencxServer) UnlockOrders() {
	server.orderMutex.Unlock()
}

// InitMatchingMaps creates the maps so we don't get nil map errors
func (server *OpencxServer) InitMatchingMaps() {
	server.orderMutex = new(sync.Mutex)
	server.OrderMap = make(map[match.Pair][]*match.LimitOrder)
}

// MatchingLoop is supposed to be run as a goroutine. This will always match stuff, and creates the server mutex map and order map
func (server *OpencxServer) MatchingLoop(pair *match.Pair, bufferSize int) {

	for {

		server.LockOrders()

		_, foundOrders := server.OrderMap[*pair]
		if foundOrders && len(server.OrderMap[*pair]) >= bufferSize {

			logging.Infof("Server order queue reached %d; Matching all prices.", bufferSize)
			server.OrderMap[*pair] = []*match.LimitOrder{}

			server.LockIngests()
			if err := server.OpencxDB.RunMatching(pair); err != nil {
				// gotta put these here cause if it errors out then oops just locked everything
				// logging.Errorf("Error with matching: \n%s", err)
			}
			server.UnlockIngests()
		}

		server.UnlockOrders()
	}

}

// InitServer creates a new server
func InitServer(db cxdb.OpencxStore, homedir string, rpcport uint16, pairsArray []*match.Pair, assets []match.Asset) *OpencxServer {
	return &OpencxServer{
		OpencxDB:           db,
		OpencxRoot:         homedir,
		OpencxPort:         rpcport,
		PairsArray:         pairsArray,
		AssetArray:         assets,
		orderMutex:         new(sync.Mutex),
		OrderMap:           make(map[match.Pair][]*match.LimitOrder),
		ingestMutex:        *new(sync.Mutex),
		BlockChanMap:       make(map[int]chan *wire.MsgBlock),
		HeightEventChanMap: make(map[int]chan lnutil.HeightEvent),
	}
}

// TODO now that I know how to use this hdkeychain stuff, let's figure out how to create addresses to store

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(privkey *[32]byte) error {

	rootBTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.TestNet3Params)
	if err != nil {
		return fmt.Errorf("Error creating master BTC Test key from private key: \n%s", err)
	}
	server.OpencxBTCTestPrivKey = rootBTCKey

	rootVTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.VertcoinRegTestParams)
	if err != nil {
		return fmt.Errorf("Error creating master VTC Test key from private key: \n%s", err)
	}
	server.OpencxVTCTestPrivKey = rootVTCKey

	rootLTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.LiteCoinTestNet4Params)
	if err != nil {
		return fmt.Errorf("Error creating master LTC Test key from private key: \n%s", err)
	}
	server.OpencxLTCTestPrivKey = rootLTCKey

	return nil
}

// SetupLitNode sets up the lit node for use later, I wanna do this because this really shouldn't be in initialization code? should it be?
// basically just run this after you unlock the key
func (server *OpencxServer) SetupLitNode(privkey *[32]byte, nodePath string, trackerURL string, proxyURL string, nat string) (err error) {
	if server.ExchangeNode, err = qln.NewLitNode(privkey, nodePath, trackerURL, proxyURL, nat); err != nil {
		return
	}

	return
}

// SetupBTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupBTCChainhook(errChan chan error, coinTypeChan chan int, hostString string) {
	var btcParam *coinparam.Params
	var err error
	var coinType int

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error when starting btc wallet: \n%s", err)
		}
		errChan <- err
		coinTypeChan <- coinType
	}()

	if lnutil.YupString(hostString) {
		btcParam = &coinparam.TestNet3Params
	} else {
		btcParam = &coinparam.RegressionNetParams
		btcParam.StartHeight = 0
	}

	btcParam.DiffCalcFunction = dummyDifficulty

	logging.Infof("Starting BTC Wallet\n")

	var btcWallet *wallit.Wallit
	if btcWallet, coinType, err = wallit.NewWallit(server.OpencxVTCTestPrivKey, btcParam.StartHeight, true, hostString, server.OpencxRoot, "", btcParam); err != nil {
		return
	}

	logging.Infof("BTC Wallet Started, cointype: %d\n", coinType)

	blockChan := btcWallet.Hook.RawBlocks()
	btcHeightChan := btcWallet.LetMeKnowHeight()
	server.OpencxBTCWallet = btcWallet
	// server.HeightEventChanMap[coinType] = btcHeightChan

	go server.HeightHandler(btcHeightChan, blockChan, btcParam)

	return
}

// SetupLTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupLTCChainhook(errChan chan error, coinTypeChan chan int, hostString string) {
	var ltcParam *coinparam.Params
	var err error
	var coinType int

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error when starting ltc wallet: \n%s", err)
		}
		errChan <- err
		coinTypeChan <- coinType
	}()

	if lnutil.YupString(hostString) {
		// TODO: move all this stuff up to be server parameters. Find a way to elegantly manage and add multiple chains while keeping track of parameters
		// and nicely connecting to nodes, while handling unable to connect stuff
		ltcParam = &coinparam.LiteCoinTestNet4Params
		ltcParam.PoWFunction = dummyProofOfWork
	} else {
		ltcParam = &coinparam.LiteRegNetParams
		ltcParam.StartHeight = 0
	}

	// difficulty in non bitcoin testnets has an air of mystery
	ltcParam.DiffCalcFunction = dummyDifficulty

	logging.Infof("Starting LTC Wallet\n")

	var ltcWallet *wallit.Wallit
	if ltcWallet, coinType, err = wallit.NewWallit(server.OpencxVTCTestPrivKey, ltcParam.StartHeight, true, hostString, server.OpencxRoot, "", ltcParam); err != nil {
		return
	}

	logging.Infof("LTC Wallet started, coinType: %d\n", coinType)

	blockChan := ltcWallet.Hook.RawBlocks()
	ltcHeightChan := ltcWallet.LetMeKnowHeight()
	server.OpencxLTCWallet = ltcWallet
	// server.HeightEventChanMap[coinType] = ltcHeightChan

	go server.HeightHandler(ltcHeightChan, blockChan, ltcParam)

	return
}

// SetupVTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupVTCChainhook(errChan chan error, coinTypeChan chan int, hostString string) {
	var vtcParam *coinparam.Params
	var err error
	var coinType int

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error when starting vtc wallet: \n%s", err)
		}
		errChan <- err
		coinTypeChan <- coinType
	}()

	if lnutil.YupString(hostString) {
		vtcParam = &coinparam.VertcoinTestNetParams
		vtcParam.PoWFunction = dummyProofOfWork
		vtcParam.DNSSeeds = []string{"jlovejoy.mit.edu", "gertjaap.ddns.net", "fr1.vtconline.org", "tvtc.vertcoin.org"}
	} else {
		vtcParam = &coinparam.VertcoinRegTestParams
	}

	vtcParam.DiffCalcFunction = dummyDifficulty

	logging.Infof("Starting VTC Wallet\n")

	var vtcWallet *wallit.Wallit
	if vtcWallet, coinType, err = wallit.NewWallit(server.OpencxVTCTestPrivKey, vtcParam.StartHeight, true, hostString, server.OpencxRoot, "", vtcParam); err != nil {
		return
	}

	logging.Infof("VTC Wallet started, coinType: %d\n", coinType)

	blockChan := vtcWallet.Hook.RawBlocks()
	vtcHeightChan := vtcWallet.LetMeKnowHeight()
	server.OpencxVTCWallet = vtcWallet
	// server.HeightEventChanMap[coinType] = vtcHeightChan

	go server.HeightHandler(vtcHeightChan, blockChan, vtcParam)

	return
}

// LinkAllWallets will link the exchanges' wallets with the lit node running.
func (server *OpencxServer) LinkAllWallets(btcCoinType int, ltcCoinType int, vtcCoinType int) (err error) {
	// Idk if I should run a tower with these, probably. It's an exchange
	if err = server.LinkOneWallet(server.OpencxBTCWallet, btcCoinType, false); err != nil {
		err = fmt.Errorf("Error linking BTC Wallet to node: \n%s", err)
		return
	}

	if err = server.LinkOneWallet(server.OpencxLTCWallet, ltcCoinType, false); err != nil {
		err = fmt.Errorf("Error linking LTC Wallet to node: \n%s", err)
		return
	}

	if err = server.LinkOneWallet(server.OpencxVTCWallet, vtcCoinType, false); err != nil {
		err = fmt.Errorf("Error linking VTC Wallet to node: \n%s", err)
		return
	}

	logging.Infof("Successfully linked all wallets!")
	return
}

// LinkOneWallet is a modified version of linkwallet in lit that doesn't make the wallet but links it with an already running one. Your responsibility to pass the correct cointype and tower.
func (server *OpencxServer) LinkOneWallet(wallet *wallit.Wallit, coinType int, tower bool) (err error) {
	// we don't need param passed as a parameter to this function, the wallet already has it so we have to substitute a bunch of stuff
	WallitIdx := wallet.Param.HDCoinType

	// see if we've already attached a wallet for this coin type
	if server.ExchangeNode.SubWallet[WallitIdx] != nil {
		err = fmt.Errorf("coin type %d already linked", WallitIdx)
		return
	}

	// see if there are other wallets already linked
	if len(server.ExchangeNode.SubWallet) != 0 {
		// there are; assert multiwallet (may already be asserted)
		server.ExchangeNode.MultiWallet = true
	}

	// Have to do this because we deleted the lines actually creating the wallet, we already have it created.
	server.ExchangeNode.SubWallet[WallitIdx] = wallet

	// if there aren't, Multiwallet will still be false; set new wallit to
	// be the first & default

	if server.ExchangeNode.ConnectedCoinTypes == nil {
		server.ExchangeNode.ConnectedCoinTypes = make(map[uint32]bool)
		server.ExchangeNode.ConnectedCoinTypes[uint32(coinType)] = true
	}

	// why is this needed in 2 places, can't this be the only time this is run?
	server.ExchangeNode.ConnectedCoinTypes[uint32(coinType)] = true

	// re-register channel addresses
	qChans, err := server.ExchangeNode.GetAllQchans()
	if err != nil {
		return err
	}

	for _, qChan := range qChans {
		var pkh [20]byte
		pkhSlice := btcutil.Hash160(qChan.MyRefundPub[:])
		copy(pkh[:], pkhSlice)
		server.ExchangeNode.SubWallet[WallitIdx].ExportHook().RegisterAddress(pkh)

		logging.Infof("Registering outpoint %v", qChan.PorTxo.Op)

		server.ExchangeNode.SubWallet[WallitIdx].WatchThis(qChan.PorTxo.Op)
	}

	go server.ExchangeNode.OPEventHandler(server.ExchangeNode.SubWallet[WallitIdx].LetMeKnow())
	go server.ExchangeNode.HeightEventHandler(server.HeightEventChanMap[coinType])

	// If this is the first coin we're linking then set that one to default.
	if !server.ExchangeNode.MultiWallet {
		server.ExchangeNode.DefaultCoin = wallet.Param.HDCoinType
	}

	// if this node is running a watchtower, link the watchtower to the
	// new wallet block events

	if tower {
		err = server.ExchangeNode.Tower.HookLink(
			server.ExchangeNode.LitFolder, wallet.Param, server.ExchangeNode.SubWallet[WallitIdx].ExportHook())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFundHandler gets the handler func to pass in to the register function. Maybe just make this a normal function, but I think we need server stuff
func (server *OpencxServer) GetFundHandler() (hFunc func(event eventbus.Event) eventbus.EventHandleResult) {
	hFunc = func(event eventbus.Event) (res eventbus.EventHandleResult) {
		// // We know this is a channel state update event
		// chanStateEvent := event.(qln.ChannelStateUpdateEvent)

		// amt := chanStateEvent.

		return
	}
	return
}

// HeightHandler is a handler for when there is a height and block event. We need both channels to work and be synchronized, which I'm assuming is the case in the lit repos. Will need to double check.
func (server *OpencxServer) HeightHandler(incomingBlockHeight chan lnutil.HeightEvent, blockChan chan *wire.MsgBlock, coinType *coinparam.Params) {
	for {

		logging.Infof("waiting for blockheight %s", coinType.Name)
		h := <-incomingBlockHeight

		logging.Infof("waiting for block %s", coinType.Name)
		block := <-blockChan
		logging.Debugf("Ingesting %d transactions at height %d\n", len(block.Transactions), h.Height)
		if err := server.ingestTransactionListAndHeight(block.Transactions, uint64(h.Height), coinType); err != nil {
			logging.Infof("something went horribly wrong with %s\n", coinType.Name)
			logging.Errorf("Here's what went horribly wrong: %s\n", err)
		}
		// logging.Infof("Passing heightevent to other chan")

		// passing heightevent to other chan
		server.HeightEventChanMap[int(coinType.HDCoinType)] <- h

	}
}

// MatchingRoutine is supposed to be run as a goroutine from placeorder so we can wait a bit while we run an order
func (server *OpencxServer) MatchingRoutine(started chan bool, pair *match.Pair, price float64) {
	server.LockIngests()
	started <- true
	if err := server.OpencxDB.RunMatchingForPrice(pair, price); err != nil {
		logging.Errorf("Error running matching while doing matching routine: \n%s")
		server.UnlockIngests()
	}

	server.UnlockIngests()
}

func dummyDifficulty(headers []*wire.BlockHeader, height int32, p *coinparam.Params) (uint32, error) {
	return headers[len(headers)-1].Bits, nil
}

func dummyProofOfWork(b []byte, height int32) chainhash.Hash {
	lowestPow := make([]byte, 32)
	smolhash, err := chainhash.NewHash(lowestPow)
	if err != nil {
		logging.Errorf("Error setting hash to a bunch of bytes of zeros")
	}
	return *smolhash
}
