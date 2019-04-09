package cxserver

import (
	"fmt"
	"sync"

	"github.com/mit-dci/lit/crypto/koblitz"

	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/uspv"

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
	OpencxPort uint16
	OpencxRoot string
	WallitRoot string
	AssetArray []match.Asset

	registrationString string
	getOrdersString    string

	ExchangeNode *qln.LitNode

	BlockChanMap       map[int]chan *wire.MsgBlock
	HeightEventChanMap map[int]chan lnutil.HeightEvent
	ingestMutex        sync.Mutex

	// All you should need to add a new coin to the exchange is the correct coin params to connect
	// to nodes and (if it works), do proof of work and such.
	HookMap    map[*coinparam.Params]*uspv.ChainHook
	WalletMap  map[*coinparam.Params]*wallit.Wallit
	PrivKeyMap map[*coinparam.Params]*hdkeychain.ExtendedKey

	// This is how we're going to easily add multiple coins
	CoinList []*coinparam.Params

	orderMutex *sync.Mutex
	OrderMap   map[match.Pair][]*match.LimitOrder

	// default Capacity is the default capacity that we send back to people.
	// remove this when we have some sense of how much money the exchange has and/or some fancy
	// algorithms to determine this number based on reputation or something
	defaultCapacity int64
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
func InitServer(db cxdb.OpencxStore, homedir string, rpcport uint16, coinList []*coinparam.Params) (server *OpencxServer) {
	server = &OpencxServer{
		OpencxDB:           db,
		OpencxPort:         rpcport,
		OpencxRoot:         homedir,
		registrationString: "opencx-register",
		getOrdersString:    "opencx-getorders",
		orderMutex:         new(sync.Mutex),
		OrderMap:           make(map[match.Pair][]*match.LimitOrder),
		ingestMutex:        *new(sync.Mutex),
		BlockChanMap:       make(map[int]chan *wire.MsgBlock),
		HeightEventChanMap: make(map[int]chan lnutil.HeightEvent),

		HookMap:    make(map[*coinparam.Params]*uspv.ChainHook),
		WalletMap:  make(map[*coinparam.Params]*wallit.Wallit),
		PrivKeyMap: make(map[*coinparam.Params]*hdkeychain.ExtendedKey),

		CoinList:        coinList,
		defaultCapacity: 1000000,
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
