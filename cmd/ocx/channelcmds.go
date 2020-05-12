package main

import (
	"fmt"

	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

var getLitConnectionCommand = &Command{
	Format: fmt.Sprintf("%s\n", lnutil.Red("getlitconnection")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Get the lightning node address and hostname in order to connect to the exchange and open up a channel.",
		"Once connected, you can push funds to the exchange as a deposit.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get the lightning node address and hostname in order to connect to the exchange and open up a channel."),
}

func (cl *ocxClient) GetLitConnection(args []string) (err error) {
	getLitConnectionReply := new(cxrpc.GetLitConnectionReply)

	if getLitConnectionReply, err = cl.RPCClient.GetLitConnection(); err != nil {
		return
	}

	for _, port := range getLitConnectionReply.Ports {
		logging.Infof("Exchange Listener: con %s@%s:%d", getLitConnectionReply.PubKeyHash, cl.RPCClient.GetHostname(), port)
	}
	return
}
