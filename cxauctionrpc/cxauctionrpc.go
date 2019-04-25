package cxauctionrpc

import "github.com/mit-dci/opencx/cxauctionserver"

// OpencxAuctionRPC is a listener for RPC commands
type OpencxAuctionRPC struct {
	Server    *cxauctionserver.OpencxAuctionServer
	OffButton chan bool
}
