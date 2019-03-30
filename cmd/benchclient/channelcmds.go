package benchclient

import (
	"github.com/mit-dci/opencx/cxrpc"
)

// GetLitConnection gets the lit con to pass in to lit. Maybe do this more automatically later on
// TODO: in order for all the trading to work properly we need to switch from names to pubkeys
func (cl *BenchClient) GetLitConnection() (getLitConnectionReply *cxrpc.GetLitConnectionReply, err error) {
	getLitConnectionReply = new(cxrpc.GetLitConnectionReply)
	getLitConnectionArgs := &cxrpc.GetLitConnectionArgs{}

	if err = cl.Call("OpencxRPC.GetLitConnection", getLitConnectionArgs, getLitConnectionReply); err != nil {
		return
	}

	return
}

// WithdrawToLightningNode takes in some arguments such as public key, amount, and ln node address
func (cl *BenchClient) WithdrawToLightningNode() (withdrawToLightningNodeReply *cxrpc.WithdrawToLightningNodeReply, err error) {
	withdrawToLightningNodeReply = new(cxrpc.WithdrawToLightningNodeReply)
	withdrawToLightningNodeArgs := &cxrpc.WithdrawToLightningNodeArgs{}

	if err = cl.Call("OpencxRPC.WithdrawToLightningNode", withdrawToLightningNodeArgs, withdrawToLightningNodeReply); err != nil {
		return
	}

	return
}
