package chainutils

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
)

// GetParamFromName gets coin params from a name
func GetParamFromName(name string) (coinType *coinparam.Params, err error) {
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

// GetParamFromHDCoinType gets coin params from a hdCoinType
func GetParamFromHDCoinType(hdCoinType uint32) (coinType *coinparam.Params, err error) {
	// create map for that O(1) access
	coinMap := map[uint32]*coinparam.Params{
		coinparam.BitcoinParams.HDCoinType:  &coinparam.BitcoinParams,
		coinparam.VertcoinParams.HDCoinType: &coinparam.VertcoinParams,
		// coinparam.LitecoinParams.HDCoinType: &coinparam.LitecoinParams,
		coinparam.TestNet3Params.HDCoinType:         &coinparam.TestNet3Params,
		coinparam.VertcoinTestNetParams.HDCoinType:  &coinparam.VertcoinTestNetParams,
		coinparam.LiteCoinTestNet4Params.HDCoinType: &coinparam.LiteCoinTestNet4Params,
		coinparam.RegressionNetParams.HDCoinType:    &coinparam.RegressionNetParams,
		coinparam.VertcoinRegTestParams.HDCoinType:  &coinparam.VertcoinRegTestParams,
		coinparam.LiteRegNetParams.HDCoinType:       &coinparam.LiteRegNetParams,
	}

	// grab from map
	var found bool
	if coinType, found = coinMap[hdCoinType]; !found {
		err = fmt.Errorf("Coin not found when trying to get from hdCoinType, maybe it's not supported yet")
		return
	}

	return
}
