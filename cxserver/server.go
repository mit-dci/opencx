package cxserver

import (
	"fmt"
	"os"
	"sync"

	"github.com/mit-dci/lit/crypto/koblitz"

	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/uspv"

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
	"github.com/mit-dci/opencx/util"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB      cxdb.OpencxStore
	OpencxRoot    string
	OpencxPort    uint16
	WallitRoot    string
	ChainhookRoot string
	LitRoot       string
	AssetArray    []match.Asset

	registrationString string
	getOrdersString    string

	ExchangeNode *qln.LitNode

	BlockChanMap       map[int]chan *wire.MsgBlock
	HeightEventChanMap map[int]chan lnutil.HeightEvent
	ingestMutex        sync.Mutex

	// These are supposed to replace the various BTC/LTC/VTC chainhooks above.
	// All you should need to add a new coin to the exchange is the correct coin params to connect
	// to nodes and (if it works), do proof of work and such.
	HookMap    map[*coinparam.Params]*uspv.ChainHook
	WalletMap  map[*coinparam.Params]*wallit.Wallit
	PrivKeyMap map[*coinparam.Params]*hdkeychain.ExtendedKey

	// This is how we're going to easily add multiple coins
	CoinList []*coinparam.Params

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

// InitServer creates a new server
func InitServer(db cxdb.OpencxStore, homedir string, rpcport uint16, coinList []*coinparam.Params) *OpencxServer {
	server := &OpencxServer{
		OpencxDB:           db,
		OpencxRoot:         homedir,
		OpencxPort:         rpcport,
		registrationString: "opencx-register",
		getOrdersString:    "opencx-getorders",
		orderMutex:         new(sync.Mutex),
		OrderMap:           make(map[match.Pair][]*match.LimitOrder),
		ingestMutex:        *new(sync.Mutex),
		BlockChanMap:       make(map[int]chan *wire.MsgBlock),
		HeightEventChanMap: make(map[int]chan lnutil.HeightEvent),
		WallitRoot:         homedir + "wallit/",
		ChainhookRoot:      homedir + "chainhook/",
		LitRoot:            homedir + "lit/",

		HookMap:    make(map[*coinparam.Params]*uspv.ChainHook),
		WalletMap:  make(map[*coinparam.Params]*wallit.Wallit),
		PrivKeyMap: make(map[*coinparam.Params]*hdkeychain.ExtendedKey),

		CoinList: coinList,
	}
	var err error
	// create wallit root directory
	_, err = os.Stat(server.WallitRoot)
	if os.IsNotExist(err) {
		err = os.Mkdir(server.WallitRoot, 0700)
	}
	if err != nil {
		logging.Errorf("Error while creating a directory: \n%s", err)
	}

	// create chainhook root directory
	_, err = os.Stat(server.ChainhookRoot)
	if os.IsNotExist(err) {
		err = os.Mkdir(server.ChainhookRoot, 0700)
	}
	if err != nil {
		logging.Errorf("Error while creating a directory: \n%s", err)
	}

	// create lit root directory
	_, err = os.Stat(server.LitRoot)
	if os.IsNotExist(err) {
		err = os.Mkdir(server.LitRoot, 0700)
	}
	if err != nil {
		logging.Errorf("Error while creating a directory: \n%s", err)
	}

	return server
}

// TODO now that I know how to use this hdkeychain stuff, let's figure out how to create addresses to store

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(privkey *[32]byte) (err error) {

	if err = server.SetupManyKeys(privkey, server.CoinList); err != nil {
		return
	}

	return nil
}

// SetupManyKeys sets up many keys for the server based on an array of coinparams.
func (server *OpencxServer) SetupManyKeys(privkey *[32]byte, paramList []*coinparam.Params) (err error) {
	for _, param := range paramList {
		if err = server.SetupSingleKey(privkey, param); err != nil {
			return
		}
	}
	return
}

// SetupSingleKey sets up a single key based on a single param for the server.
func (server *OpencxServer) SetupSingleKey(privkey *[32]byte, param *coinparam.Params) (err error) {
	var rootKey *hdkeychain.ExtendedKey
	if rootKey, err = hdkeychain.NewMaster(privkey[:], param); err != nil {
		err = fmt.Errorf("Error creating master %s key from private key: \n%s", param.Name, err)
		return
	}
	server.PrivKeyMap[param] = rootKey

	return
}

// SetupLitNode sets up the lit node for use later, I wanna do this because this really shouldn't be in initialization code? should it be?
// basically just run this after you unlock the key
func (server *OpencxServer) SetupLitNode(privkey *[32]byte, trackerURL string, proxyURL string, nat string) (err error) {
	if server.ExchangeNode, err = qln.NewLitNode(privkey, server.LitRoot, trackerURL, proxyURL, nat); err != nil {
		return
	}

	return
}

// SetupWallet sets up a wallet for a specific coin, based on params.
func (server *OpencxServer) SetupWallet(errChan chan error, param *coinparam.Params, resync bool, hostString string) {
	var err error
	var coinType int
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error when starting wallet: \n%s", err)
		}
		errChan <- err
	}()

	logging.Infof("Starting %s wallet\n", param.Name)

	key, found := server.PrivKeyMap[param]
	if !found {
		err = fmt.Errorf("Could not find key for wallet. Aborting wallet setup")
		return
	}

	var wallet *wallit.Wallit
	if wallet, coinType, err = wallit.NewWallit(key, param.StartHeight, resync, hostString, server.WallitRoot, "", param); err != nil {
		return
	}

	server.WalletMap[param] = wallet

	logging.Infof("%s wallet Started, cointype: %d\n", param.Name, coinType)
	// figure out whether or not to do this if merged

	server.StartChainhookHandlers(wallet)

	return
}

// SetupAllWallets sets up all wallets with parameters as specified in the hostParamList
func (server *OpencxServer) SetupAllWallets(hostParamList util.HostParamList, resync bool) (err error) {
	hpLen := len(hostParamList)
	errChan := make(chan error, hpLen)
	for _, hostParam := range hostParamList {
		go server.SetupWallet(errChan, hostParam.Param, resync, hostParam.Host)
	}

	for i := 0; i < hpLen; i++ {
		if err = <-errChan; err != nil {
			return
		}
	}
	return
}

// StartChainhookHandlers gets the channels from the wallet's chainhook and starts a handler.
func (server *OpencxServer) StartChainhookHandlers(wallet *wallit.Wallit) {
	hook := wallet.ExportHook()

	server.HookMap[wallet.Param] = &hook

	hookBlockChan := hook.NewRawBlocksChannel()
	currentHeightChan := hook.NewHeightChannel()

	logging.Infof("Successfully set up chainhook from wallet, starting handlers")
	go server.ChainHookHeightHandler(currentHeightChan, hookBlockChan, wallet.Param)

}

// LinkAllWallets will link the exchanges' wallets with the lit node running. Defaults to false for running tower.
func (server *OpencxServer) LinkAllWallets() (err error) {

	// Not sure whether or not this should just assume that everything in the map is what you want, but I'm going to
	// assume that if there's a coin / param in the CoinList that isn't in the wallet map, then the wallets haven't
	// been started or something is wrong. This is definitely a synchronous thing to be doing, you need to start
	// the wallets for all your coins before you try to link them all. If you don't want to link them all, use
	// LinkManyWallets.
	for _, param := range server.CoinList {
		wallet, found := server.WalletMap[param]
		if !found {
			err = fmt.Errorf("Wallet in Coin List not being tracked by exchange in map, start it please")
		}

		// Idk if I should run a tower with these, probably. It's an exchange
		if err = server.LinkOneWallet(wallet, false); err != nil {
			return
		}

	}

	logging.Infof("Successfully linked all wallets!")
	return
}

// LinkManyWallets takes in a bunch of wallets and links them. We set the tower for all wallets consistently
// coinType as specified in the parameters.
func (server *OpencxServer) LinkManyWallets(wallets []*wallit.Wallit, tower bool) (err error) {
	for _, wallet := range wallets {
		if err = server.LinkOneWallet(wallet, tower); err != nil {
			return
		}
	}

	return
}

// LinkOneWallet is a modified version of linkwallet in lit that doesn't make the wallet but links it with an already running one. Your responsibility to pass the correct cointype and tower.
func (server *OpencxServer) LinkOneWallet(wallet *wallit.Wallit, tower bool) (err error) {
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
	}

	// why is this needed in 2 places, can't this be the only time this is run?
	server.ExchangeNode.ConnectedCoinTypes[WallitIdx] = true

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
	go server.ExchangeNode.HeightEventHandler(server.HeightEventChanMap[int(WallitIdx)])

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

// GetRegistrationString gets a string that should be signed in order for a client to be registered
func (server *OpencxServer) GetRegistrationString() (regStr string) {
	regStr = server.registrationString
	return
}

// RegistrationStringVerify verifies a signature for a registration string and returns a pubkey
func (server *OpencxServer) RegistrationStringVerify(sig []byte) (pubkey *koblitz.PublicKey, err error) {
	// e = h(registrationstring)
	sha3 := sha3.New256()
	sha3.Write([]byte(server.GetRegistrationString()))
	e := sha3.Sum(nil)

	if pubkey, _, err = koblitz.RecoverCompact(koblitz.S256(), sig, e); err != nil {
		err = fmt.Errorf("Error verifying registration string, invalid signature: \n%s", err)
		return
	}

	return
}

// GetOrdersString gets a string that should be signed in order for a client to be registered
func (server *OpencxServer) GetOrdersString() (getOrderStr string) {
	getOrderStr = server.getOrdersString
	return
}

// GetOrdersStringVerify verifies a signature for the getOrdersString
func (server *OpencxServer) GetOrdersStringVerify(sig []byte) (pubkey *koblitz.PublicKey, err error) {
	// e = h(getOrders)
	sha3 := sha3.New256()
	sha3.Write([]byte(server.GetOrdersString()))
	e := sha3.Sum(nil)

	if pubkey, _, err = koblitz.RecoverCompact(koblitz.S256(), sig, e); err != nil {
		err = fmt.Errorf("Error verifying getOrders string, invalid signature: \n%s", err)
		return
	}

	return
}

// GetAddressMap gets an address map for a pubkey. This is so we can register multiple ways.
func (server *OpencxServer) GetAddressMap(pubkey *koblitz.PublicKey) (addrMap map[*coinparam.Params]string, err error) {
	// go through each enabled wallet in the server and create a new address for them.
	addrMap = make(map[*coinparam.Params]string)
	for param := range server.WalletMap {
		if addrMap[param], err = server.GetAddrForCoin(param, pubkey); err != nil {
			return
		}
	}
	return
}
