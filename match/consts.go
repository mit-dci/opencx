// Package match provides utilities and useful structures for building exchange systems.
package match

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	util "github.com/mit-dci/opencx/chainutils"
)

const (
	// BTC is a constant used to represent a BTC token
	BTC Asset = 0x00
	// VTC is a constant used to represent a VTC token
	VTC Asset = 0x01

	// Wait until you have a LTC coinparam
	// LTC is a constant used to represent a LTC token
	// LTC Asset = 0x02

	// BTCTest is a constant used to represent a BTC Test net token
	BTCTest Asset = 0x03
	// VTCTest is a constant used to represent a VTC Test net token
	VTCTest Asset = 0x04
	// LTCTest is a constant used to represent a LTC Test net token
	LTCTest Asset = 0x05
	// BTCReg is a constant used to represent a BTC Reg net token
	BTCReg Asset = 0x06
	// VTCReg is a constant used to represent a VTC Reg net token
	VTCReg Asset = 0x07
	// LTCReg is a constant used to represent a LTC Reg net token
	LTCReg Asset = 0x08
)

// largeAssetList is something used for testing the generateassetpairs function, this should be put into a unit test once tests are written
func largeAssetList() []Asset {
	return []Asset{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}
}

// AssetFromCoinParam gets a byte representation of an asset from a coinparam
func AssetFromCoinParam(cpm *coinparam.Params) (a Asset, err error) {

	// create map for that O(1) access
	assetMap := map[*coinparam.Params]Asset{
		&coinparam.BitcoinParams:  BTC,
		&coinparam.VertcoinParams: VTC,
		// &coinparam.LitecoinParams: LTC,
		&coinparam.TestNet3Params:         BTCTest,
		&coinparam.VertcoinTestNetParams:  VTCTest,
		&coinparam.LiteCoinTestNet4Params: LTCTest,
		&coinparam.RegressionNetParams:    BTCReg,
		&coinparam.VertcoinRegTestParams:  VTCReg,
		&coinparam.LiteRegNetParams:       LTCReg,
	}

	// grab from map
	var found bool
	if a, found = assetMap[cpm]; !found {
		err = fmt.Errorf("Could not get an asset for that coin param")
		return
	}

	return
}

// CoinParamFromAsset is the reverse of AssetFromCoinParam. TODO: change the coinparam so it has an inherent ID anyways?
func (a Asset) CoinParamFromAsset() (coinType *coinparam.Params, err error) {
	// create map for that O(1) access
	assetMap := map[Asset]*coinparam.Params{
		BTC: &coinparam.BitcoinParams,
		VTC: &coinparam.VertcoinParams,
		// LTC: &coinparam.LitecoinParams,
		BTCTest: &coinparam.TestNet3Params,
		VTCTest: &coinparam.VertcoinTestNetParams,
		LTCTest: &coinparam.LiteCoinTestNet4Params,
		BTCReg:  &coinparam.RegressionNetParams,
		LTCReg:  &coinparam.LiteRegNetParams,
		VTCReg:  &coinparam.VertcoinRegTestParams,
	}

	// grab from map
	var found bool
	if coinType, found = assetMap[a]; !found {
		err = fmt.Errorf("Could not get a coin param for that asset")
		return
	}

	return

}

// AssetFromString returns an asset from a string
func AssetFromString(name string) (a Asset, err error) {
	// we do this to enforce consistency sorta. I don't want this asset thing and the coin params that we support to be separated. They are right now, though.
	// the main reason they're separated is that I want the coin params to have their own unique byte so the client only sends a byte to indicate which asset
	// they want. Like an asset ID. TODO: Change this to HDCoinType, or magic bytes in the future. We need an asset id.
	var cpm *coinparam.Params
	if cpm, err = util.GetParamFromName(name); err != nil {
		return
	}

	if a, err = AssetFromCoinParam(cpm); err != nil {
		return
	}

	return
}

func (a Asset) String() string {
	var err error
	var coinType *coinparam.Params
	if coinType, err = a.CoinParamFromAsset(); err != nil {
		return "UnknownAsset"
	}
	return coinType.Name
}
