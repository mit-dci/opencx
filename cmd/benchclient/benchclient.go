package benchclient

import (
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxrpc"
)

// BenchClient holds the RPC Client and defines many methods that can be called
type BenchClient struct {
	hostname  string
	port      uint16
	RPCClient cxrpc.OpencxClient
	PrivKey   *koblitz.PrivateKey
}

// SetupBenchClient creates a new BenchClient for use as an RPC Client
func (cl *BenchClient) SetupBenchClient(server string, port uint16) (err error) {
	cl.RPCClient = new(cxrpc.OpencxRPCClient)
	cl.hostname = server
	cl.port = port

	// we set the privkey here because we aren't using a command line to send orders
	if err = cl.RPCClient.SetupConnection(server, port); err != nil {
		return
	}

	return
}

// SetupBenchNoiseClient create a new BenchClient for use as an RPC-Noise Client
func (cl *BenchClient) SetupBenchNoiseClient(server string, port uint16) (err error) {

	// Authenticate with the same key that BenchClient uses for signatures
	noiseClient := new(cxrpc.OpencxNoiseClient)
	if err = noiseClient.SetKey(cl.PrivKey); err != nil {
		return
	}

	// Now that the key is set we can start doing stuff.
	cl.RPCClient = noiseClient
	cl.hostname = server
	cl.port = port

	// we set the privkey here because we aren't using a command line to send orders
	if err = cl.RPCClient.SetupConnection(server, port); err != nil {
		return
	}

	return
}

// Call calls a method from the rpc client
func (cl *BenchClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}

// GetHostname returns the hostname, used for convenience I guess, maybe move out of benchclient and into ocx?
func (cl *BenchClient) GetHostname() string {
	return cl.hostname
}

// GetPort returns the port, used for convenience I guess, maybe move out of benchclient and into ocx?
func (cl *BenchClient) GetPort() uint16 {
	return cl.port
}
