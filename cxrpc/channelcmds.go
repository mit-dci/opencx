package cxrpc

import (
	"strconv"
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

	for _, port := range ports {
		// TODO: figure out how to use the rest of the port list
		var port64 uint64
		port64, err = strconv.ParseUint(port, 10, 16)
		reply.Ports = append(reply.Ports, uint16(port64))
	}

	return
}
