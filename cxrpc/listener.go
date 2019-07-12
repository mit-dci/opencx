package cxrpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxnoise"
	"github.com/mit-dci/opencx/logging"
)

// NoiseListen is a synchronous version of RPCListenAsync
func NoiseListen(rpc1 *OpencxRPC, privkey *koblitz.PrivateKey, host string, port uint16) {

	doneChan := make(chan bool, 1)
	go NoiseListenAsync(doneChan, privkey, rpc1, host, port)
	<-doneChan

	return
}

// NoiseListenAsync listens on socket host and port
func NoiseListenAsync(doneChan chan bool, privkey *koblitz.PrivateKey, rpc1 *OpencxRPC, host string, port uint16) {
	var err error

	// Start noise rpc server (need to do this since the client is a rpc newclient)
	noiseRPCServer := rpc.NewServer()

	logging.Infof("Registering RPC API over Noise protocol ...")
	// Register rpc
	if err = noiseRPCServer.Register(rpc1); err != nil {
		logging.Fatalf("Error registering RPC Interface:\n%s", err)
	}

	logging.Infof("Starting RPC Server over noise protocol")
	// Start RPC Server
	var listener net.Listener
	if listener, err = cxnoise.NewListener(privkey, int(port)); err != nil {
		logging.Fatal("listen error:", err)
	}
	logging.Infof("Running RPC-Noise server on %s\n", listener.Addr().String())

	// We don't need to do anything fancy here either because the noise protocol
	// is built in to the listener as well.
	go noiseRPCServer.Accept(listener)
	go OffButtonCloseListener(rpc1, listener)
	doneChan <- true
	return
}

// RPCListen is a synchronous version of RPCListenAsync
func RPCListen(rpc1 *OpencxRPC, host string, port uint16) {

	doneChan := make(chan bool, 1)
	go RPCListenAsync(doneChan, rpc1, host, port)
	<-doneChan

	return
}

// RPCListenAsync listens on socket host and port
func RPCListenAsync(doneChan chan bool, rpc1 *OpencxRPC, host string, port uint16) {
	var err error

	logging.Infof("Registering RPC API...")
	// Register rpc
	if err = rpc.Register(rpc1); err != nil {
		logging.Fatalf("Error registering RPC Interface:\n%s", err)
	}

	logging.Infof("Starting RPC Server")
	// Start RPC Server
	serverAddr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	var listener net.Listener
	if listener, err = net.Listen("tcp", serverAddr); err != nil {
		logging.Fatal("listen error:", err)
	}
	logging.Infof("Running RPC server on %s\n", listener.Addr().String())

	go rpc.Accept(listener)

	go OffButtonCloseListener(rpc1, listener)
	doneChan <- true
	return
}

// OffButtonCloseListener waits for the off button to close the listener
func OffButtonCloseListener(rpc1 *OpencxRPC, listener net.Listener) {
	for {
		<-rpc1.OffButton
		logging.Infof("Got stop request, closing tcp listener")
		if err := listener.Close(); err != nil {
			logging.Errorf("Error closing listener: \n%s", err)
		}
		return
	}
}
