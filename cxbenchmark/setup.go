package cxbenchmark

import (
	"fmt"

	"github.com/btcsuite/golangcrypto/sha3"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/cxserver"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = uint16(12347)
)

func SetupBenchmarkClientWithKey(clientPrivKey *koblitz.PrivateKey) (client *benchclient.BenchClient, err error) {
	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchClient(defaultServer, defaultPort); err != nil {
		err = fmt.Errorf("Error setting up OpenCX RPC Client: \n%s", err)
		return
	}
	return
}

// SetupBenchmarkClient sets up the benchmark and returns the client
func SetupBenchmarkClient() (client *benchclient.BenchClient, err error) {

	// TODO: Figure out why this is here and remove if unnecessary, uncomment if necessary
	// logging.SetLogLevel(3)

	// have to set this for non noise client because while we don't use things for authentication we do use it for signing
	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Error setting key for client: \n%s", err)
		return
	}

	if client, err = SetupBenchmarkClientWithKey(clientPrivKey); err != nil {
		err = fmt.Errorf("Error setting up benchmark client with key: %s", err)
		return
	}

	return
}

// SetupNoiseBenchmarkClient sets up the benchmark and returns the client
func SetupNoiseBenchmarkClient() (client *benchclient.BenchClient, err error) {

	// TODO: Figure out why this is here and remove if unnecessary, uncomment if necessary
	// logging.SetLogLevel(3)

	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Error setting key for noise client: \n%s", err)
		return
	}

	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchNoiseClient(defaultServer, defaultPort); err != nil {
		err = fmt.Errorf("Error setting up OpenCX RPC-Noise Client: \n%s", err)
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

// setupEasyBenchmarkDualClient sets up an environment where we can test two infinitely funded clients
func setupEasyBenchmarkDualClient(authrpc bool) (client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, offChan chan bool, err error) {
	var serverKey *koblitz.PrivateKey
	if serverKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Could not get new private key: \n%s", err)
		return
	}

	var clientKeyOne *koblitz.PrivateKey
	if clientKeyOne, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Could not get new client #1 privkey: %s", err)
		return
	}

	var clientKeyTwo *koblitz.PrivateKey
	if clientKeyTwo, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		err = fmt.Errorf("Could not get new client #2 privkey: %s", err)
		return
	}

	whitelist := []*koblitz.PublicKey{
		clientKeyOne.PubKey(),
		clientKeyTwo.PubKey(),
	}

	// set the closing function here as well
	// We don't really care about the actual server object if it's running and we know it's running
	if _, offChan, err = createDefaultLightServerWithKey(serverKey, whitelist, authrpc); err != nil {
		err = fmt.Errorf("Could not create default server with key: \n%s", err)
	}

	if client1, err = SetupBenchmarkClientWithKey(clientKeyOne); err != nil {
		err = fmt.Errorf("Error setting up benchmark client for client 1: \n%s", err)
		return
	}

	if client2, err = SetupBenchmarkClientWithKey(clientKeyTwo); err != nil {
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

	return
}
