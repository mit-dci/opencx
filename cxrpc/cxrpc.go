package cxrpc

import (
	"net"

	"github.com/mit-dci/opencx/cxserver"
)

// OpencxRPC is what is registered and called
type OpencxRPC struct {
	Server *cxserver.OpencxServer
}

// OpencxRPCCaller is a listener for RPC commands
type OpencxRPCCaller struct {
	caller   *OpencxRPC
	listener net.Listener
	killers  []chan bool
}
