package cxbenchmark

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/benchclient"
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

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

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

	if ocxServer, err = cxserver.InitServer(setEngines, mengines, limBooks, depositStores, setStores, ""); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Check that the private key exists and if it does, load it
	if err = ocxServer.SetupServerKeys(key); err != nil {
		err = fmt.Errorf("Error setting up server keys: \n%s", err)
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

// createFullServer creates a server after starting the database with a bunch of parameters
func createFullServer(coinList []*coinparam.Params, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxserver.OpencxServer, offChan chan bool, err error) {

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Could not generate asset pairs from coin list: %s", err)
		return
	}

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

	if ocxServer, err = cxserver.InitServer(setEngines, mengines, limBooks, depositStores, setStores, ""); err != nil {
		err = fmt.Errorf("Error initializing server for createFullServer: %s", err)
		return
	}

	key := new([32]byte)
	copy(key[:], privkey.Serialize())

	// Check that the private key exists and if it does, load it
	if err = ocxServer.SetupServerKeys(key); err != nil {
		err = fmt.Errorf("Error setting up server keys: \n%s", err)
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

// createDefaultParamServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultParamServerWithKey(privkey *koblitz.PrivateKey, authrpc bool) (server *cxserver.OpencxServer, offChan chan bool, err error) {
	return createFullServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, "localhost", uint16(12346), privkey, authrpc)
}

// createDefaultLightServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultLightServerWithKey(privkey *koblitz.PrivateKey, whitelist []*koblitz.PublicKey, authrpc bool) (server *cxserver.OpencxServer, offChan chan bool, err error) {
	return createLightSettleServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, whitelist, "localhost", uint16(12346), privkey, authrpc)
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
