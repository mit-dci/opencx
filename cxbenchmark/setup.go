package cxbenchmark

import (
	"log"

	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = 12345
)

// SetupBenchmark sets up the benchmark and returns the client
func SetupBenchmark() *benchclient.BenchClient {
	var err error

	logging.SetLogLevel(2)

	client := new(benchclient.BenchClient)
	if err = client.SetupBenchClient(defaultServer, defaultPort); err != nil {
		log.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
	}

	return client
}
