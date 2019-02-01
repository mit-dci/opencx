package match

import (
	"fmt"

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

func (d *Deposit) String() string {
	return fmt.Sprintf("Deposit: {\n\tName: %s\n\tAddress: %s\n\tAmount: %d\n\tTxid: %s\n\tCoinType: %s\n\tBlockHeightReceived: %d\n\tConfirmations: %d\n}",
		d.Name, d.Address, d.Amount, d.Txid, d.CoinType.Name, d.BlockHeightReceived, d.Confirmations)
}
