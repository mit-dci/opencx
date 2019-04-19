package cxbenchmark

import (
	"log"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = uint16(12345)
)

// SetupBenchmark sets up the benchmark and returns the client
func SetupBenchmark() (client *benchclient.BenchClient) {
	var err error

	logging.SetLogLevel(3)

	// have to set this for non noise client because while we don't use thigns for authentication we do use it for signing
	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		log.Fatalf("Error setting key for client: \n%s", err)
	}

	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchClient(defaultServer, defaultPort); err != nil {
		log.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
	}

	return
}

// SetupNoiseBenchmark sets up the benchmark and returns the client
func SetupNoiseBenchmark() (client *benchclient.BenchClient) {
	var err error

	logging.SetLogLevel(3)

	var clientPrivKey *koblitz.PrivateKey
	if clientPrivKey, err = koblitz.NewPrivateKey(koblitz.S256()); err != nil {
		log.Fatalf("Error setting key for client: \n%s", err)
	}

	client = &benchclient.BenchClient{
		PrivKey: clientPrivKey,
	}

	if err = client.SetupBenchNoiseClient(defaultServer, defaultPort); err != nil {
		log.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
	}

	return
}
