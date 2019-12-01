package cxserver

import (
	"fmt"
	"sync"

	"github.com/Rjected/lit/crypto/koblitz"

	"golang.org/x/crypto/sha3"

	"github.com/Rjected/lit/uspv"

	"github.com/Rjected/lit/btcutil/hdkeychain"
	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/lnutil"
	"github.com/Rjected/lit/qln"
	"github.com/Rjected/lit/wallit"
	"github.com/Rjected/lit/wire"

	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// OpencxServer is what orchestrates the exchange. It's where you plug everything into basically.
// The Server looks spookily like a node.
type OpencxServer struct {
	SettlementEngines map[*coinparam.Params]match.SettlementEngine
	MatchingEngines   map[match.Pair]match.LimitEngine
	Orderbooks        map[match.Pair]match.LimitOrderbook
	DepositStores     map[*coinparam.Params]cxdb.DepositStore
	SettlementStores  map[*coinparam.Params]cxdb.SettlementStore
	dbLock            *sync.Mutex

	registrationString string
	getOrdersString    string

	ExchangeNode *qln.LitNode

	BlockChanMap       map[int]chan *wire.MsgBlock
	HeightEventChanMap map[int]chan lnutil.HeightEvent
	ingestMutex        sync.Mutex

	OpencxRoot string

	// All you should need to add a new coin to the exchange is the correct coin params to connect
	// to nodes and (if it works), do proof of work and such.
	HookMap    map[*coinparam.Params]*uspv.ChainHook
	hookMtx    *sync.Mutex
	WalletMap  map[*coinparam.Params]*wallit.Wallit
	walletMtx  *sync.Mutex
	PrivKeyMap map[*coinparam.Params]*hdkeychain.ExtendedKey
	privKeyMtx *sync.Mutex

	// default Capacity is the default capacity that we send back to people.
	// remove this when we have some sense of how much money the exchange has and/or some fancy
	// algorithms to determine this number based on reputation or something
	defaultCapacity int64
}

// InitServer creates a new server
func InitServer(setEngines map[*coinparam.Params]match.SettlementEngine, matchEngines map[match.Pair]match.LimitEngine, books map[match.Pair]match.LimitOrderbook, depositStores map[*coinparam.Params]cxdb.DepositStore, settleStores map[*coinparam.Params]cxdb.SettlementStore, rootDir string) (server *OpencxServer, err error) {
	server = &OpencxServer{
		SettlementEngines: setEngines,
		MatchingEngines:   matchEngines,
		Orderbooks:        books,
		DepositStores:     depositStores,
		SettlementStores:  settleStores,
		dbLock:            new(sync.Mutex),
		OpencxRoot:        rootDir,

		registrationString: "opencx-register",
		getOrdersString:    "opencx-getorders",
		ingestMutex:        *new(sync.Mutex),
		BlockChanMap:       make(map[int]chan *wire.MsgBlock),
		HeightEventChanMap: make(map[int]chan lnutil.HeightEvent),

		HookMap:    make(map[*coinparam.Params]*uspv.ChainHook),
		WalletMap:  make(map[*coinparam.Params]*wallit.Wallit),
		PrivKeyMap: make(map[*coinparam.Params]*hdkeychain.ExtendedKey),

		hookMtx:    new(sync.Mutex),
		walletMtx:  new(sync.Mutex),
		privKeyMtx: new(sync.Mutex),

		defaultCapacity: 1000000,
	}

	return
}

// StartChainhookHandlers gets the channels from the wallet's chainhook and starts a handler.
func (server *OpencxServer) StartChainhookHandlers(wallet *wallit.Wallit) {
	hook := wallet.ExportHook()

	server.hookMtx.Lock()
	server.HookMap[wallet.Param] = &hook
	server.hookMtx.Unlock()

	hookBlockChan := hook.NewRawBlocksChannel()
	currentHeightChan := hook.NewHeightChannel()

	logging.Infof("Successfully set up chainhook from wallet, starting handlers")
	go server.ChainHookHeightHandler(currentHeightChan, hookBlockChan, wallet.Param)

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

// GetPairs just iterates throug the matching engine map, getting their pair keys and appending
// to a list
func (server *OpencxServer) GetPairs() (pairs []*match.Pair) {
	server.dbLock.Lock()
	var currPair *match.Pair
	for pair, _ := range server.MatchingEngines {
		currPair = new(match.Pair)
		*currPair = pair
		pairs = append(pairs, currPair)
	}
	for _, p := range pairs {
		logging.Infof("pair: %s", p.PrettyString())
	}
	server.dbLock.Unlock()
	return
}
