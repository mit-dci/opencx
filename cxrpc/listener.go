package cxrpc

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/lndc"
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

	logging.Infof("Registering RPC API over Noise protocol ...")
	// Register rpc
	if err = rpc.Register(rpc1); err != nil {
		logging.Fatalf("Error registering RPC Interface:\n%s", err)
	}

	logging.Infof("Starting RPC Server over noise protocol")
	// Start RPC Server
	var listener net.Listener
	if listener, err = lndc.NewListener(privkey, int(port)); err != nil {
		logging.Fatal("listen error:", err)
	}
	logging.Infof("Running RPC-Noise server on %s\n", listener.Addr().String())

	OffButtonCloseListener(rpc1, listener)
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
	var listener net.Listener
	if listener, err = net.Listen("tcp", host+":"+fmt.Sprintf("%d", port)); err != nil {
		logging.Fatal("listen error:", err)
	}
	logging.Infof("Running RPC server on %s\n", listener.Addr().String())

	rpc.HandleHTTP()
	go http.Serve(listener, nil)

	OffButtonCloseListener(rpc1, listener)
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
