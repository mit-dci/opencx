package cxserver

import (
	"fmt"
	"time"

	"github.com/mit-dci/lit/litrpc"
	"github.com/mit-dci/opencx/logging"
)

// SetupLitRPCConnect sets up an rpc connection with a running lit node?
func (server *OpencxServer) SetupLitRPCConnect(rpchost string, rpcport uint16) {
	var err error
	defer func() {
		if err != nil {
			logging.Errorf("Error creating lit RPC connection: \n%s", err)
		}
	}()
	if server.ExchangeNode == nil {
		err = fmt.Errorf("Please start the exchange node before trying to create its RPC")
		return
	}

	rpc1 := new(litrpc.LitRPC)
	rpc1.Node = server.ExchangeNode
	rpc1.OffButton = make(chan bool, 1)
	server.ExchangeNode.RPC = rpc1

	// we don't care about unauthRPC
	go litrpc.RPCListen(rpc1, rpchost, rpcport)

	<-rpc1.OffButton
	logging.Infof("Got stop request\n")
	time.Sleep(time.Second)
	return
}
