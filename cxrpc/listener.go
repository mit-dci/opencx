package cxrpc

import (
	"fmt"
	"net"
	"net/rpc"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxnoise"
	"github.com/mit-dci/opencx/cxserver"
	"github.com/mit-dci/opencx/logging"
)

func CreateRPCForServer(server *cxserver.OpencxServer) (rpc1 *OpencxRPCCaller, err error) {
	rpc1 = &OpencxRPCCaller{
		caller: &OpencxRPC{
			Server: server,
		},
	}
	return
}

// NoiseListen is a synchronous version of RPCListenAsync
func (rpc1 *OpencxRPCCaller) NoiseListen(privkey *koblitz.PrivateKey, host string, port uint16) (err error) {

	doneChan := make(chan bool, 1)
	errChan := make(chan error, 1)
	go rpc1.NoiseListenAsync(doneChan, errChan, privkey, host, port)
	select {
	case err = <-errChan:
	case <-doneChan:
	}

	return
}

// NoiseListenAsync listens on socket host and port
func (rpc1 *OpencxRPCCaller) NoiseListenAsync(doneChan chan bool, errChan chan error, privkey *koblitz.PrivateKey, host string, port uint16) {
	var err error
	if rpc1.caller == nil {
		errChan <- fmt.Errorf("Error, rpc caller cannot be nil, please create caller correctly")
		close(errChan)
		return
	}

	// Start noise rpc server (need to do this since the client is a rpc newclient)
	noiseRPCServer := rpc.NewServer()

	logging.Infof("Registering RPC API over Noise protocol ...")
	// Register rpc
	if err = noiseRPCServer.Register(rpc1.caller); err != nil {
		errChan <- fmt.Errorf("Error registering RPC Interface: %s", err)
		close(errChan)
		return
	}

	logging.Infof("Starting RPC Server over noise protocol")
	// Start RPC Server
	if rpc1.listener, err = cxnoise.NewListener(privkey, int(port)); err != nil {
		errChan <- fmt.Errorf("Error creating noise listener for NoiseListenAsync: %s", err)
		close(errChan)
		return
	}
	logging.Infof("Running RPC-Noise server on %s\n", rpc1.listener.Addr().String())

	// We don't need to do anything fancy here either because the noise protocol
	// is built in to the listener as well.
	go noiseRPCServer.Accept(rpc1.listener)
	doneChan <- true
	close(doneChan)
	return
}

// RPCListen is a synchronous version of RPCListenAsync
func (rpc1 *OpencxRPCCaller) RPCListen(host string, port uint16) (err error) {

	doneChan := make(chan bool, 1)
	errChan := make(chan error, 1)
	go rpc1.RPCListenAsync(doneChan, errChan, host, port)
	select {
	case err = <-errChan:
	case <-doneChan:
	}

	return
}

// RPCListenAsync listens on socket host and port
func (rpc1 *OpencxRPCCaller) RPCListenAsync(doneChan chan bool, errChan chan error, host string, port uint16) {
	var err error
	if rpc1.caller == nil {
		errChan <- fmt.Errorf("Error, rpc caller cannot be nil, please create caller correctly")
		close(errChan)
		return
	}

	logging.Infof("Registering RPC API...")
	// Register rpc
	if err = rpc.Register(rpc1.caller); err != nil {
		errChan <- fmt.Errorf("Error registering RPC Interface: %s", err)
		close(errChan)
		return
	}

	logging.Infof("Starting RPC Server")
	// Start RPC Server
	serverAddr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	if rpc1.listener, err = net.Listen("tcp", serverAddr); err != nil {
		errChan <- fmt.Errorf("Error listening for RPCListenAsync: %s", err)
		close(errChan)
		return
	}
	logging.Infof("Running RPC server on %s\n", rpc1.listener.Addr().String())

	go rpc.Accept(rpc1.listener)
	doneChan <- true
	close(doneChan)
	return
}

// WaitUntilDead waits until the Stop() method is called
func (rpc1 *OpencxRPCCaller) WaitUntilDead() {
	dedchan := make(chan bool, 1)
	rpc1.killers = append(rpc1.killers, dedchan)
	<-dedchan
	return
}

// Stop closes the RPC listener and notifies those from WaitUntilDead
func (rpc1 *OpencxRPCCaller) Stop() (err error) {
	if rpc1.listener == nil {
		err = fmt.Errorf("Error, cannot stop a listener that doesn't exist")
		return
	}
	logging.Infof("Stopping RPC!!")
	if err = rpc1.listener.Close(); err != nil {
		err = fmt.Errorf("Error closing listener: %s", err)
		return
	}
	// kill the guy waiting
	for _, killer := range rpc1.killers {
		// send the signals, but even if they don't send, close the channel
		select {
		case killer <- true:
			close(killer)
		default:
			close(killer)
		}
	}
	return
}
