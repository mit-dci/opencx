package match

import (
	"github.com/mit-dci/lit/coinparam"
)

// Deposit is a struct that represents a deposit on chain
type Deposit struct {
	Name                string
	Address             string
	Amount              uint64
	Txid                string
	CoinType            *coinparam.Params
	BlockHeightReceived uint64
	Confirmations       uint64
}
