package cxbenchmark

import (
	"fmt"
	"time"

	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"
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

var (
	newVTC = changePort(&coinparam.VertcoinRegTestParams, "20444")
)

// changePort is a gross method that I wrote. It's a hack to change the port of the VertcoinRegTestParams because the Bitcoin regtest and Vertcoin regtest ports are the same
func changePort(param *coinparam.Params, port string) (newparam *coinparam.Params) {
	newparam = param
	newparam.DefaultPort = port
	return
}

// createLightSettleServer creates a server with "pinky swear settlement" after starting the database with a bunch of parameters for everything else
func createLightSettleServer(coinList []*coinparam.Params, whitelist []*koblitz.PublicKey, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (rpcListener *cxrpc.OpencxRPCCaller, err error) {

	// logging.SetLogLevel(3)
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
	if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(wlMap, false); err != nil {
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
	var ocxServer *cxserver.OpencxServer
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
	if rpcListener, err = cxrpc.CreateRPCForServer(ocxServer); err != nil {
		err = fmt.Errorf("Error creating rpc for server in createLightSettleServer: %s", err)
		return
	}

	if !authrpc {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for server: %s", err)
			return
		}
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for server: %s", err)
			return
		}
	}

	// // Register RPC Commands and set server
	// rpc1 := new(cxrpc.OpencxRPC)
	// rpc1.OffButton = make(chan bool, 1)
	// rpc1.Server = ocxServer

	// if !authrpc {
	// 	// this tells us when the rpclisten is done
	// 	doneChan := make(chan bool, 1)
	// 	logging.Infof(" === will start to listen on rpc ===")
	// 	go cxrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	// } else {

	// 	privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
	// 	// this tells us when the rpclisten is done
	// 	doneChan := make(chan bool, 1)
	// 	logging.Infof(" === will start to listen on noise-rpc ===")
	// 	go cxrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	// }

	// offChan = rpc1.OffButton
	return
}

// createLightAuctionServer creates a server with "pinky swear settlement" after starting the database with a bunch of parameters for everything else
func createLightAuctionServer(coinList []*coinparam.Params, whitelist []*koblitz.PublicKey, maxBatchSize uint64, auctionTime uint64, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (rpcListener *cxauctionrpc.AuctionRPCCaller, err error) {

	logging.Infof("Create light auction server start -- %s", time.Now())

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
	if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(wlMap, false); err != nil {
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
	var ocxServer *cxauctionserver.OpencxAuctionServer
	if ocxServer, err = cxauctionserver.InitServer(setEngines, mengines, aucBooks, pzEngines, batchers, 100, auctionTime); err != nil {
		err = fmt.Errorf("Error initializing server for createLightAuctionServer: %s", err)
		return
	}

	if err = ocxServer.StartClockRandomAuction(); err != nil {
		err = fmt.Errorf("Error starting clock: %s", err)
		return
	}

	logging.Infof("Starting RPC Listen process.")
	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Register RPC Commands and set server
	if rpcListener, err = cxauctionrpc.CreateRPCForServer(ocxServer); err != nil {
		err = fmt.Errorf("Error creating rpc for server in createLightAuctionServer: %s", err)
		return
	}

	if !authrpc {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for auction server: %s", err)
			return
		}
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for auction server: %s", err)
			return
		}
	}

	return
}

// createFullServer creates a server after starting the database with a bunch of parameters
func createFullServer(coinList []*coinparam.Params, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (rpcListener *cxrpc.OpencxRPCCaller, err error) {

	// logging.SetLogLevel(3)
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
	var ocxServer *cxserver.OpencxServer
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
	if rpcListener, err = cxrpc.CreateRPCForServer(ocxServer); err != nil {
		err = fmt.Errorf("Error creating rpc for server in createLightSettleServer: %s", err)
		return
	}

	if !authrpc {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for server: %s", err)
			return
		}
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for server: %s", err)
			return
		}
	}

	// // Register RPC Commands and set server
	// rpc1 := new(cxrpc.OpencxRPC)
	// rpc1.OffButton = make(chan bool, 1)
	// rpc1.Server = ocxServer

	// if !authrpc {
	// 	// this tells us when the rpclisten is done
	// 	doneChan := make(chan bool, 1)
	// 	logging.Infof(" === will start to listen on rpc ===")
	// 	go cxrpc.RPCListenAsync(doneChan, rpc1, serverhost, serverport)
	// 	// wait till rpc stuff is done
	// 	<-doneChan
	// 	close(doneChan)
	// } else {

	// 	privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
	// 	// this tells us when the rpclisten is done
	// 	doneChan := make(chan bool, 1)
	// 	logging.Infof(" === will start to listen on noise-rpc ===")
	// 	go cxrpc.NoiseListenAsync(doneChan, privkey, rpc1, serverhost, serverport)
	// 	// wait till rpc stuff is done
	// 	<-doneChan
	// 	close(doneChan)
	// }

	// offChan = rpc1.OffButton

	return
}

// createFullAuctionServer creates an auction server after starting the database with a bunch of parameters
func createFullAuctionServer(coinList []*coinparam.Params, maxBatchSize uint64, auctionTime uint64, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (rpcListener *cxauctionrpc.AuctionRPCCaller, err error) {

	logging.Infof("Create full auction server start -- %s", time.Now())

	// defaults -- orderChanSize is like 100, TODO: delete orderChanSize because it's probably obsolete
	var ocxServer *cxauctionserver.OpencxAuctionServer
	if ocxServer, err = cxauctionserver.InitServerSQLDefault(coinList, 100, auctionTime, maxBatchSize); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	if err = ocxServer.StartClockRandomAuction(); err != nil {
		err = fmt.Errorf("Error starting clock: %s", err)
		return
	}

	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Register RPC Commands and set server
	if rpcListener, err = cxauctionrpc.CreateRPCForServer(ocxServer); err != nil {
		err = fmt.Errorf("Error creating rpc for server in createLightAuctionServer: %s", err)
		return
	}

	if !authrpc {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for auction server: %s", err)
			return
		}
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, serverhost, serverport); err != nil {
			err = fmt.Errorf("Error listening for RPC for auction server: %s", err)
			return
		}
	}

	return
}

// createDefaultParamServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultParamServerWithKey(privkey *koblitz.PrivateKey, authrpc bool) (rpcListener *cxrpc.OpencxRPCCaller, err error) {
	return createFullServer([]*coinparam.Params{&coinparam.RegressionNetParams, newVTC, &coinparam.LiteRegNetParams}, "localhost", uint16(12347), privkey, authrpc)
}

// createDefaultLightServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultLightServerWithKey(privkey *koblitz.PrivateKey, whitelist []*koblitz.PublicKey, authrpc bool) (rpcListener *cxrpc.OpencxRPCCaller, err error) {
	return createLightSettleServer([]*coinparam.Params{&coinparam.RegressionNetParams, newVTC, &coinparam.LiteRegNetParams}, whitelist, "localhost", uint16(12347), privkey, authrpc)
}

// createDefaultLightAuctionServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultLightAuctionServerWithKey(privkey *koblitz.PrivateKey, whitelist []*koblitz.PublicKey, authrpc bool) (rpcListener *cxauctionrpc.AuctionRPCCaller, err error) {
	return createLightAuctionServer([]*coinparam.Params{&coinparam.RegressionNetParams, newVTC, &coinparam.LiteRegNetParams}, whitelist, 100, 7000000, "localhost", uint16(12347), privkey, authrpc)
}
