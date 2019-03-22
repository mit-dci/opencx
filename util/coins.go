package util

import (
	"github.com/mit-dci/lit/btcutil/chaincfg"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wire"
)

// GetSchemaNameFromCoinType is a function that returns a string according to how the schema is laid out.
func GetSchemaNameFromCoinType(coinType *coinparam.Params) (schemaName string, err error) {
	schemaName = coinType.Name
	return
}

// ConvertChaincfgCoinparam just converts from coinparam to chaincfg so we can do btcutil stuff.
// Would be ideal if we just had all the btcutil stuff attached to coinparam from the beginning.
func ConvertChaincfgCoinparam(p *coinparam.Params) (param *chaincfg.Params) {
	param = &chaincfg.Params{
		Name:                     p.Name,
		Net:                      wire.BitcoinNet(p.NetMagicBytes),
		DefaultPort:              p.DefaultPort,
		DNSSeeds:                 p.DNSSeeds,
		GenesisBlock:             p.GenesisBlock,
		PowLimit:                 p.PowLimit,
		PowLimitBits:             p.PowLimitBits,
		CoinbaseMaturity:         p.CoinbaseMaturity,
		SubsidyReductionInterval: p.SubsidyReductionInterval,
		TargetTimespan:           p.TargetTimespan,
		TargetTimePerBlock:       p.TargetTimePerBlock,
		RetargetAdjustmentFactor: p.RetargetAdjustmentFactor,
		ReduceMinDifficulty:      p.ReduceMinDifficulty,
		MinDiffReductionTime:     p.MinDiffReductionTime,
		GenerateSupported:        p.GenerateSupported,
		Checkpoints:              ConvertCheckpointArray(p.Checkpoints),
		BlockEnforceNumRequired:  p.BlockEnforceNumRequired,
		BlockRejectNumRequired:   p.BlockRejectNumRequired,
		BlockUpgradeNumToCheck:   p.BlockUpgradeNumToCheck,
		RelayNonStdTxs:           p.RelayNonStdTxs,
		PubKeyHashAddrID:         p.PubKeyHashAddrID,
		ScriptHashAddrID:         p.ScriptHashAddrID,
		PrivateKeyID:             p.PrivateKeyID,
		Bech32Prefix:             p.Bech32Prefix,
		HDPrivateKeyID:           p.HDPrivateKeyID,
		HDPublicKeyID:            p.HDPublicKeyID,
		HDCoinType:               p.HDCoinType,
		// We haven't dealt with forks yet so ForkID isn't a thing yet
	}

	return
}

// ConvertCheckpointArray converts a coinparam checkpoint array into a chaincfg checkpoint array
func ConvertCheckpointArray(cpmCheckpointArray []coinparam.Checkpoint) (cfgCheckpointArray []chaincfg.Checkpoint) {
	var checkpoints []chaincfg.Checkpoint
	for _, checkpoint := range cpmCheckpointArray {
		newCheckpoint := ConvertCheckpoint(checkpoint)
		checkpoints = append(checkpoints, newCheckpoint)
	}
	return
}

// ConvertCheckpoint converts a coinparam checkpoint into a chaincfg checkpoint
func ConvertCheckpoint(cpmCheckpoint coinparam.Checkpoint) (cfgCheckpoint chaincfg.Checkpoint) {
	return chaincfg.Checkpoint{
		Height: cpmCheckpoint.Height,
		Hash:   cpmCheckpoint.Hash,
	}
}

// GetCoinTypeFromName gets coin params from a name
func GetCoinTypeFromName(name string) (coinType *coinparam.Params) {
	// TODO

	return
}
