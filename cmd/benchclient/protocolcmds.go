package benchclient

import "github.com/mit-dci/opencx/cxauctionrpc"

// GetPublicParameters returns the public parameters like the auction time and current auction ID
func (cl *BenchClient) GetPublicParameters() (getPublicParametersReply *cxauctionrpc.GetPublicParametersReply, err error) {
	getPublicParametersReply = new(cxauctionrpc.GetPublicParametersReply)
	getPublicParametersArgs := new(cxauctionrpc.GetPublicParametersArgs)

	// Actually use the RPC Client to call the method
	if err = cl.Call("OpencxAuctionRPC.GetPublicParameters", getPublicParametersArgs, getPublicParametersReply); err != nil {
		return
	}

	return
}
