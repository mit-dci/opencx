package benchclient

import (
	"github.com/mit-dci/opencx/cxrpc"
)

// BenchClient holds the RPC Client and defines many methods that can be called
type BenchClient struct {
	RPCClient *cxrpc.OpencxRPCClient
	PrivKey   *[32]byte
}

// SetupBenchClient creates a new BenchClient for use as an RPC Client
func (cl *BenchClient) SetupBenchClient(server string, port uint16) error {
	var err error

	cl = &BenchClient{
		RPCClient: new(cxrpc.OpencxRPCClient),
	}

	// we set the privkey here because we aren't using a command line to send orders
	err = cl.RPCClient.SetupConnection(server, port)
	if err != nil {
		return err
	}

	return nil
}

// Call calls a method from the rpc client
func (cl *BenchClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}
