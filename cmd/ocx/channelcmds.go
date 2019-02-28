package main

import (
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

func (cl *openCxClient) GetLitConnection(args []string) (err error) {
	getLitConnectionReply := new(cxrpc.GetLitConnectionReply)

	if getLitConnectionReply, err = cl.RPCClient.GetLitConnection(); err != nil {
		return
	}

	for _, port := range getLitConnectionReply.Ports {
		logging.Infof("Exchange Listener: con %s@%s:%d", getLitConnectionReply.PubKeyHash, cl.RPCClient.GetHostname(), port)
	}
	return
}
