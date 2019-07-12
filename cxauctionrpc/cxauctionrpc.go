package cxauctionrpc

import (
	"net"

	"github.com/mit-dci/opencx/cxauctionserver"
)

// AuctionRPCCaller is a listener for RPC commands
type AuctionRPCCaller struct {
	caller   *OpencxAuctionRPC
	listener net.Listener
	killers  []chan bool
}

// OpencxAuctionRPC is what is registered and called
type OpencxAuctionRPC struct {
	Server *cxauctionserver.OpencxAuctionServer
}
