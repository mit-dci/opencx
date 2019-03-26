package util

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
)

// GetSchemaNameFromCoinType is a function that returns a string according to how the schema is laid out.
func GetSchemaNameFromCoinType(coinType *coinparam.Params) (schemaName string, err error) {
	schemaName = coinType.Name
	return
}

// GetCoinTypeFromName gets coin params from a name
func GetCoinTypeFromName(name string) (coinType *coinparam.Params, err error) {
	// create map for that O(1) access
	coinMap := map[string]*coinparam.Params{
		coinparam.BitcoinParams.Name:  &coinparam.BitcoinParams,
		coinparam.VertcoinParams.Name: &coinparam.VertcoinParams,
		// coinparam.LitecoinParams.Name: &coinparam.LitecoinParams,
		coinparam.TestNet3Params.Name:         &coinparam.TestNet3Params,
		coinparam.VertcoinTestNetParams.Name:  &coinparam.VertcoinTestNetParams,
		coinparam.LiteCoinTestNet4Params.Name: &coinparam.LiteCoinTestNet4Params,
		coinparam.RegressionNetParams.Name:    &coinparam.RegressionNetParams,
		coinparam.VertcoinRegTestParams.Name:  &coinparam.VertcoinRegTestParams,
		coinparam.LiteRegNetParams.Name:       &coinparam.LiteRegNetParams,
	}

	// grab from map
	var found bool
	if coinType, found = coinMap[name]; !found {
		err = fmt.Errorf("Coin not found when trying to get from name, maybe it's not supported yet")
		return
	}

	return
}
