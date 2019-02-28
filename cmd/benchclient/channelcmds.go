package benchclient

import (
	"github.com/mit-dci/opencx/cxrpc"
)

// GetLitConnection gets the lit con to pass in to lit. Maybe do this more automatically later on
// TODO: in order for all the trading to work properly we need to switch from names to pubkeys
func (cl *BenchClient) GetLitConnection() (getLitConnectionReply *cxrpc.GetLitConnectionReply, err error) {
	getLitConnectionArgs := &cxrpc.GetLitConnectionArgs{}

	if err = cl.Call("OpencxRPC.SubmitOrder", getLitConnectionArgs, getLitConnectionReply); err != nil {
		return
	}

	return
}
