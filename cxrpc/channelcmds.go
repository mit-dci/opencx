package cxrpc

import (
	"fmt"
	"net"
	"strconv"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// GetLitConnectionArgs holds the args for the getlitconnection RPC command
type GetLitConnectionArgs struct {
	// empty
}

// GetLitConnectionReply holds the reply for the getlitconnection RPC command
type GetLitConnectionReply struct {
	PubKeyHash string
	Ports      []uint16
}

// GetLitConnection gets a pubkeyhash and port for connecting with lit, the hostname is assumed to be the same.
func (cl *OpencxRPC) GetLitConnection(args GetLitConnectionArgs, reply *GetLitConnectionReply) (err error) {
	var hosts []string
	reply.PubKeyHash, hosts = cl.Server.ExchangeNode.GetLisAddressAndPorts()

	if len(hosts) == 0 {
		err = fmt.Errorf("Exchange not listening at the moment, sorry")
		return
	}
	logging.Infof("Sending connection to client, addr: %s len hosts: %d", reply.PubKeyHash, len(hosts))

	reply.Ports = make([]uint16, len(hosts))
	for i, hostport := range hosts {
		var port string
		// we don't care about the host
		if _, port, err = net.SplitHostPort(hostport); err != nil {
			return
		}

		var port64 uint64
		if port64, err = strconv.ParseUint(port, 10, 16); err != nil {
			return
		}

		reply.Ports[i] = uint16(port64)
	}

	return
}

// WithdrawToLightningNodeArgs holds the args for the withdrawtolightning RPC command
type WithdrawToLightningNodeArgs struct {
	Withdrawal *match.Withdrawal
	Signature  []byte
}

// WithdrawToLightningNodeReply holds the reply for the withdrawtolightning RPC command
type WithdrawToLightningNodeReply struct {
	Txid string
}

// WithdrawToLightningNode creates a channel that pushes a certain amount to a lightning node through a lightning channel.
func (cl *OpencxRPC) WithdrawToLightningNode(args WithdrawToLightningNodeArgs, reply *WithdrawToLightningNodeReply) (err error) {

	return
}
