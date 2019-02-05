package util

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
)

// GetSchemaNameFromCoinType is a function that returns a string according to how the schema is laid out.
// Should change the entire thing to be based off of cointype name.
func GetSchemaNameFromCoinType(coinType *coinparam.Params) (string, error) {
	if coinType == &coinparam.TestNet3Params || coinType == &coinparam.RegressionNetParams {
		return "btc", nil
	}
	if coinType == &coinparam.VertcoinTestNetParams || coinType == &coinparam.VertcoinRegTestParams {
		return "vtc", nil
	}
	if coinType == &coinparam.LiteCoinTestNet4Params || coinType == &coinparam.LiteRegNetParams {
		return "ltc", nil
	}
	return "", fmt.Errorf("Could not determine schema name from coin type with name %s", coinType.Name)
}
