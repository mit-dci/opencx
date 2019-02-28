package cxrpc

import (
	"fmt"
	"strconv"

	"github.com/mit-dci/opencx/logging"
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
	var ports []string
	reply.PubKeyHash, ports = cl.Server.ExchangeNode.GetLisAddressAndPorts()

	if len(ports) == 0 {
		err = fmt.Errorf("Exchange not listening on any ports at the moment, sorry")
		return
	}
	logging.Infof("We are in getlitconnection, addr: %s len ports: %d", reply.PubKeyHash, len(ports))

	reply.Ports = make([]uint16, len(ports))
	for _, port := range ports {
		// TODO: figure out how to use the rest of the port list
		var port64 uint64
		port64, err = strconv.ParseUint(port, 10, 16)
		reply.Ports = append(reply.Ports, uint16(port64))
		logging.Infof("Port sent to client: %d", port64)
	}

	return
}
