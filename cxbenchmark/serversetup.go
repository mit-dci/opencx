package cxbenchmark

import (
	"fmt"
	"time"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/benchclient"
	util "github.com/mit-dci/opencx/chainutils"
	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/cxauctionserver"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbmemory"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/cxserver"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// createLightSettleServer creates a server with "pinky swear settlement" after starting the database with a bunch of parameters for everything else
func createLightSettleServer(coinList []*coinparam.Params, whitelist []*koblitz.PublicKey, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxserver.OpencxServer, offChan chan bool, err error) {

	logging.SetLogLevel(3)
	logging.Infof("Create light settle server start -- %s", time.Now())

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

	logging.Infof("Creating engines...")
	var mengines map[match.Pair]match.LimitEngine
	if mengines, err = cxdbsql.CreateLimitEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating limit engine map with coinlist for createFullServer: %s", err)
		return
	}

	// These lines are the only difference between the LightSettleServer and the FullServer
	wlMap := createWhitelistMap(coinList, whitelist)
	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(wlMap); err != nil {
		err = fmt.Errorf("Error creating pinky swear settlement engine map for createFullServer: %s", err)
		return
	}

	logging.Infof("Creating stores...")
	var limBooks map[match.Pair]match.LimitOrderbook
	if limBooks, err = cxdbsql.CreateLimitOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating limit orderbook map for createFullServer: %s", err)
		return
	}

	var depositStores map[*coinparam.Params]cxdb.DepositStore
	if depositStores, err = cxdbsql.CreateDepositStoreMap(coinList); err != nil {
		err = fmt.Errorf("Error creating deposit store map for createFullServer: %s", err)
		return
	}

	var setStores map[*coinparam.Params]cxdb.SettlementStore
	if setStores, err = cxdbsql.CreateSettlementStoreMap(coinList); err != nil {
		err = fmt.Errorf("Error creating settlement store map for createFullServer: %s", err)
		return
	}

	// TODO: change this root directory nonsense!!!
	if ocxServer, err = cxserver.InitServer(setEngines, mengines, limBooks, depositStores, setStores, ".benchmarkInfo/"); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	logging.Infof("Starting RPC Listen process.")
	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Check that the private key exists and if it does, load it
	if err = ocxServer.SetupServerKeys(key); err != nil {
		err = fmt.Errorf("Error setting up server keys: \n%s", err)
	}

	// Generate the host param list
	// the host params are all of the coinparams / coins we support
	// this coinparam list is generated from the configuration file with generateHostParams
	hpList := util.HostParamsFromCoinList(coinList)

	// Set up all chain hooks and wallets
	if err = ocxServer.SetupAllWallets(hpList, "wallit/", false); err != nil {
		logging.Fatalf("Error setting up wallets: \n%s", err)
		return
	}

	// Register RPC Commands and set server
	rpc1 := new(cxrpc.OpencxRPC)
	rpc1.OffButton = make(chan bool, 1)
	rpc1.Server = ocxServer

	if !authrpc {
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on rpc ===")
		go cxrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on noise-rpc ===")
		go cxrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	}

	offChan = rpc1.OffButton
	return
}

// createLightAuctionServer creates a server with "pinky swear settlement" after starting the database with a bunch of parameters for everything else
func createLightAuctionServer(coinList []*coinparam.Params, whitelist []*koblitz.PublicKey, maxBatchSize uint64, auctionTime uint64, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxauctionserver.OpencxAuctionServer, offChan chan bool, err error) {

	logging.Infof("Create light settle server start -- %s", time.Now())

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

	logging.Infof("Creating engines...")
	var mengines map[match.Pair]match.AuctionEngine
	if mengines, err = cxdbsql.CreateAuctionEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction engine map with coinlist for createLightAuctionServer: %s", err)
		return
	}

	// These lines are the only difference between the LightAuctionServer and the FullAuctionServer
	wlMap := createWhitelistMap(coinList, whitelist)
	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(wlMap); err != nil {
		err = fmt.Errorf("Error creating pinky swear settlement engine map for createLightAuctionServer: %s", err)
		return
	}

	logging.Infof("Creating stores...")
	var aucBooks map[match.Pair]match.AuctionOrderbook
	if aucBooks, err = cxdbsql.CreateAuctionOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating auction orderbook map for createLightAuctionServer: %s", err)
		return
	}

	var pzEngines map[match.Pair]cxdb.PuzzleStore
	if pzEngines, err = cxdbsql.CreatePuzzleStoreMap(pairList); err != nil {
		err = fmt.Errorf("Error creating puzzle store map for createLightAuctionServer: %s", err)
		return
	}

	var batchers map[match.Pair]match.AuctionBatcher
	if batchers, err = cxauctionserver.CreateAuctionBatcherMap(pairList, maxBatchSize); err != nil {
		err = fmt.Errorf("Error creating batcher map for createLightAuctionServer: %s", err)
		return
	}

	// orderChanSize = 100 because uh why not?
	if ocxServer, err = cxauctionserver.InitServer(setEngines, mengines, aucBooks, pzEngines, batchers, 100, auctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	logging.Infof("Starting RPC Listen process.")
	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Register RPC Commands and set server
	rpc1 := new(cxauctionrpc.OpencxAuctionRPC)
	rpc1.OffButton = make(chan bool, 1)
	rpc1.Server = ocxServer

	if !authrpc {
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on rpc ===")
		go cxauctionrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on noise-rpc ===")
		go cxauctionrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	}

	offChan = rpc1.OffButton
	return
}

// createFullServer creates a server after starting the database with a bunch of parameters
func createFullServer(coinList []*coinparam.Params, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxserver.OpencxServer, offChan chan bool, err error) {

	logging.SetLogLevel(3)
	logging.Infof("Create full server start -- %s", time.Now())

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

	logging.Infof("Creating engines...")
	var mengines map[match.Pair]match.LimitEngine
	if mengines, err = cxdbsql.CreateLimitEngineMap(pairList); err != nil {
		err = fmt.Errorf("Error creating limit engine map with coinlist for createFullServer: %s", err)
		return
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbsql.CreateSettlementEngineMap(coinList); err != nil {
		err = fmt.Errorf("Error creating settlement engine map for createFullServer: %s", err)
		return
	}

	logging.Infof("Creating stores...")
	var limBooks map[match.Pair]match.LimitOrderbook
	if limBooks, err = cxdbsql.CreateLimitOrderbookMap(pairList); err != nil {
		err = fmt.Errorf("Error creating limit orderbook map for createFullServer: %s", err)
		return
	}

	var depositStores map[*coinparam.Params]cxdb.DepositStore
	if depositStores, err = cxdbsql.CreateDepositStoreMap(coinList); err != nil {
		err = fmt.Errorf("Error creating deposit store map for createFullServer: %s", err)
		return
	}

	var setStores map[*coinparam.Params]cxdb.SettlementStore
	if setStores, err = cxdbsql.CreateSettlementStoreMap(coinList); err != nil {
		err = fmt.Errorf("Error creating settlement store map for createFullServer: %s", err)
		return
	}

	// TODO: get rid of this directory nonsense, just figure out a nice way to deal with these things
	if ocxServer, err = cxserver.InitServer(setEngines, mengines, limBooks, depositStores, setStores, ".benchmarkInfo/"); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	logging.Infof("Starting RPC Listen process.")
	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Check that the private key exists and if it does, load it
	if err = ocxServer.SetupServerKeys(key); err != nil {
		err = fmt.Errorf("Error setting up server keys: \n%s", err)
	}

	// Generate the host param list
	// the host params are all of the coinparams / coins we support
	// this coinparam list is generated from the configuration file with generateHostParams
	hpList := util.HostParamsFromCoinList(coinList)

	// Set up all chain hooks and wallets
	if err = ocxServer.SetupAllWallets(hpList, "wallit/", false); err != nil {
		logging.Fatalf("Error setting up wallets: \n%s", err)
		return
	}

	// Register RPC Commands and set server
	rpc1 := new(cxrpc.OpencxRPC)
	rpc1.OffButton = make(chan bool, 1)
	rpc1.Server = ocxServer

	if !authrpc {
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on rpc ===")
		go cxrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on noise-rpc ===")
		go cxrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	}

	offChan = rpc1.OffButton

	return
}

// createFullAuctionServer creates an auction server after starting the database with a bunch of parameters
func createFullAuctionServer(coinList []*coinparam.Params, maxBatchSize uint64, auctionTime uint64, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxauctionserver.OpencxAuctionServer, offChan chan bool, err error) {

	logging.Infof("Create full auction server start -- %s", time.Now())

	// defaults -- orderChanSize is like 100, TODO: delete orderChanSize because it's probably obsolete
	if ocxServer, err = cxauctionserver.InitServerSQLDefault(coinList, 100, auctionTime, maxBatchSize); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Register RPC Commands and set server
	rpc1 := new(cxauctionrpc.OpencxAuctionRPC)
	rpc1.OffButton = make(chan bool, 1)
	rpc1.Server = ocxServer

	if !authrpc {
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on rpc ===")
		go cxauctionrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on noise-rpc ===")
		go cxauctionrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	}

	offChan = rpc1.OffButton

	return
}

// createDefaultParamServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultParamServerWithKey(privkey *koblitz.PrivateKey, authrpc bool) (server *cxserver.OpencxServer, offChan chan bool, err error) {
	return createFullServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, "localhost", uint16(12347), privkey, authrpc)
}

// createDefaultLightServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultLightServerWithKey(privkey *koblitz.PrivateKey, whitelist []*koblitz.PublicKey, authrpc bool) (server *cxserver.OpencxServer, offChan chan bool, err error) {
	return createLightSettleServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, whitelist, "localhost", uint16(12347), privkey, authrpc)
}

// createDefaultLightAuctionServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultLightAuctionServerWithKey(privkey *koblitz.PrivateKey, whitelist []*koblitz.PublicKey, authrpc bool) (server *cxauctionserver.OpencxAuctionServer, offChan chan bool, err error) {
	return createLightAuctionServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, whitelist, 100, 1000, "localhost", uint16(12347), privkey, authrpc)
}

// prepareBalances adds tons of money to both accounts
func prepareBalances(client *benchclient.BenchClient, server *cxserver.OpencxServer) (err error) {

	if err = server.DebitUser(client.PrivKey.PubKey(), 1000000000, &coinparam.RegressionNetParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of regtest to client: \n%s", err)
		return
	}

	if err = server.DebitUser(client.PrivKey.PubKey(), 1000000000, &coinparam.LiteRegNetParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of litereg to client: \n%s", err)
		return
	}

	if err = server.DebitUser(client.PrivKey.PubKey(), 1000000000, &coinparam.VertcoinRegTestParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of vtcreg to client: \n%s", err)
		return
	}

	return
}
