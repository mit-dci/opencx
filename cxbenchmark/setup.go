package cxbenchmark

import (
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/cxserver"
	"github.com/mit-dci/opencx/logging"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = uint16(12346)
)

// SetupBenchmarkClient sets up the benchmark and returns the client
func SetupBenchmarkClient() (client *benchclient.BenchClient, err error) {

	logging.SetLogLevel(3)

	// have to set this for non noise client because while we don't use things for authentication we do use it for signing
	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Error setting key for client: \n%s", err)
		return
	}

	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchClient(defaultServer, defaultPort); err != nil {
		err = fmt.Errorf("Error setting up OpenCX RPC Client: \n%s", err)
		return
	}

	return
}

// SetupNoiseBenchmarkClient sets up the benchmark and returns the client
func SetupNoiseBenchmarkClient() (client *benchclient.BenchClient, err error) {

	logging.SetLogLevel(3)

	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Error setting key for client: \n%s", err)
		return
	}

	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchNoiseClient(defaultServer, defaultPort); err != nil {
		err = fmt.Errorf("Error setting up OpenCX RPC Client: \n%s", err)
		return
	}

	return
}

// signBytes is used in the register method because that's an interactive process.
// BenchClient shouldn't be responsible for interactive stuff, just providing a good
// Go API for the RPC methods the exchange offers.
func signBytes(client *benchclient.BenchClient, bytes []byte) (signature []byte, err error) {

	sha := sha3.New256()
	sha.Write(bytes)
	e := sha.Sum(nil)

	if signature, err = koblitz.SignCompact(koblitz.S256(), client.PrivKey, e, false); err != nil {
		err = fmt.Errorf("Failed to sign bytes : \n%s", err)
		return
	}

	return
}

func registerClient(client *benchclient.BenchClient) (err error) {
	// Register the clients
	var regStringReply *cxrpc.GetRegistrationStringReply
	if regStringReply, err = client.GetRegistrationString(); err != nil {
		return
	}

	var sig []byte
	if sig, err = signBytes(client, []byte(regStringReply.RegistrationString)); err != nil {
		return
	}

	// we don't really care about the reply
	if _, err = client.Register(sig); err != nil {
		return
	}

	return
}

// createFullServer creates a server after starting the database with a bunch of parameters
func createFullServer(coinList []*coinparam.Params, serverhost string, serverport uint16, privkey *koblitz.PrivateKey, authrpc bool) (ocxServer *cxserver.OpencxServer, offChan chan bool, err error) {

	// Create db connection
	var db *cxdbsql.DB
	db = new(cxdbsql.DB)

	// Setup DB Client
	if err = db.SetupClient(coinList); err != nil {
		err = fmt.Errorf("Error setting up sql client: \n%s", err)
		return
	}

	// Anyways, here's where we set the server
	ocxServer = cxserver.InitServer(db, "", serverport, coinList)

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
	// // create function and add it to closing stack that we made ourselves.
	// closeListeners := func() (err error) {
	// 	// lol just turn it off
	// 	rpc1.OffButton <- true
	// 	return
	// }
	// functionCloser(closeFunc, closeListeners)

	return
}

// functionCloser modifies the original function to add another one on to its stack. Uses function pointers. Honestly just having fun here to see what compiles, might not even work. This is the most impure thing I've ever written.
func functionCloser(original func() error, additionalCloser func() error) {
	original = func() (newErr error) {
		err := original()
		newErr = additionalCloser()
		if err != nil {
			if newErr != nil {
				newErr = fmt.Errorf("Two errors: error when calling original function: %s\nOriginal errors: %s", err, newErr)
				return
			}
			err = fmt.Errorf("One original error: %s", err)
			return
		}
		if newErr != nil {
			newErr = fmt.Errorf("New error after calling additional closer: \n%s", newErr)
			return
		}
		return
	}
	return
}

// createDefaultParamServerWithKey creates a server with a bunch of default params minus privkey and authrpc
func createDefaultParamServerWithKey(privkey *koblitz.PrivateKey, authrpc bool) (server *cxserver.OpencxServer, offChan chan bool, err error) {
	return createFullServer([]*coinparam.Params{&coinparam.RegressionNetParams, &coinparam.VertcoinRegTestParams, &coinparam.LiteRegNetParams}, "localhost", uint16(12346), privkey, authrpc)
}

// prepareBalances adds tons of money to both accounts
func prepareBalances(client *benchclient.BenchClient, server *cxserver.OpencxServer) (err error) {

	if err = server.OpencxDB.AddToBalance(client.PrivKey.PubKey(), 1000000000, &coinparam.RegressionNetParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of regtest to client: \n%s", err)
		return
	}

	if err = server.OpencxDB.AddToBalance(client.PrivKey.PubKey(), 1000000000, &coinparam.LiteRegNetParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of litereg to client: \n%s", err)
		return
	}

	if err = server.OpencxDB.AddToBalance(client.PrivKey.PubKey(), 1000000000, &coinparam.VertcoinRegTestParams); err != nil {
		err = fmt.Errorf("Error adding a bunch of vtcreg to client: \n%s", err)
		return
	}

	return
}

// setupBenchmarkDualClient sets up an environment where we can test two very well funded clients client1 and client2 with configurable authrpc
func setupBenchmarkDualClient(authrpc bool) (client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, offChan chan bool, err error) {
	var serverKey *koblitz.PrivateKey
	if serverKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Could not get new private key: \n%s", err)
		return
	}

	// set the closing function here as well
	var server *cxserver.OpencxServer
	if server, offChan, err = createDefaultParamServerWithKey(serverKey, authrpc); err != nil {
		err = fmt.Errorf("Could not create default server with key: \n%s", err)
	}

	if client1, err = SetupBenchmarkClient(); err != nil {
		err = fmt.Errorf("Error setting up benchmark client for client 1: \n%s", err)
		return
	}
	if client2, err = SetupBenchmarkClient(); err != nil {
		err = fmt.Errorf("Error setting up benchmark cient for client 2: \n%s", err)
	}

	if err = registerClient(client1); err != nil {
		err = fmt.Errorf("Could not register client2: \n%s", err)
		return
	}

	if err = registerClient(client2); err != nil {
		err = fmt.Errorf("Could not register client2: \n%s", err)
		return
	}

	if err = prepareBalances(client1, server); err != nil {
		err = fmt.Errorf("Could not add balances to client1: \n%s", err)
		return
	}

	if err = prepareBalances(client2, server); err != nil {
		err = fmt.Errorf("Could not add balances to client2: \n%s", err)
		return
	}

	return
}
