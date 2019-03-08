package match

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/lit/coinparam"
)

// Deposit is a struct that represents a deposit on chain
type Deposit struct {
	Pubkey              *koblitz.PublicKey
	Address             string
	Amount              uint64
	Txid                string
	CoinType            *coinparam.Params
	BlockHeightReceived uint64
	Confirmations       uint64
}

func (d *Deposit) String() string {
	return fmt.Sprintf("Deposit: {\n\tPubkey: %x\n\tAddress: %s\n\tAmount: %d\n\tTxid: %s\n\tCoinType: %s\n\tBlockHeightReceived: %d\n\tConfirmations: %d\n}",
		d.Pubkey.SerializeCompressed(), d.Address, d.Amount, d.Txid, d.CoinType.Name, d.BlockHeightReceived, d.Confirmations)
}

// LightningDeposit is a struct that represents a deposit made with lightning
type LightningDeposit struct {
	Pubkey   *koblitz.PublicKey // maybe switch to real pubkey later
	Amount   uint64
	CoinType *coinparam.Params
	ChanIdx  uint32
}

func (ld *LightningDeposit) String() string {
	return fmt.Sprintf("Deposit: {\n\tPubkey: %x\n\tAmount: %d\n\tCoinType: %s\n\tChannelIdx: %d\n\t\n}", ld.Pubkey.SerializeCompressed(), ld.Amount, ld.CoinType.Name, ld.ChanIdx)
}
