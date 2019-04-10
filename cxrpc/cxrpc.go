package cxrpc

import "github.com/mit-dci/opencx/cxserver"

// OpencxRPC is a listener for RPC commands
type OpencxRPC struct {
	Server    *cxserver.OpencxServer
	OffButton chan bool
}
