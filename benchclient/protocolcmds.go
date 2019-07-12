package benchclient

import (
	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/match"
)

// GetPublicParameters returns the public parameters like the auction time and current auction ID
func (cl *BenchClient) GetPublicParameters(pair *match.Pair) (getPublicParametersReply *cxauctionrpc.GetPublicParametersReply, err error) {
	getPublicParametersReply = new(cxauctionrpc.GetPublicParametersReply)
	getPublicParametersArgs := &cxauctionrpc.GetPublicParametersArgs{
		Pair: *pair,
	}

	// Actually use the RPC Client to call the method
	if err = cl.Call("OpencxAuctionRPC.GetPublicParameters", getPublicParametersArgs, getPublicParametersReply); err != nil {
		return
	}

	return
}
