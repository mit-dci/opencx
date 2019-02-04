package match

import (
	"github.com/mit-dci/lit/coinparam"
)

// Pair is a struct that represents a trading pair
type Pair struct {
	// AssetWant is the asset that buyers want, and that sellers are selling. credit buyers with this.
	AssetWant asset `json:"assetWant"`
	// AssetHave is the asset that sellers want, and that buyersr are buying. credit sellers with this.
	AssetHave asset `json:"assetHave"`
}

// type aliases are only usually used for codebase refactoring, so make this better when you have time. At some point a struct will probably need to be made.
// like I'm probably going to replace this with just a master list of all the chainhooks / coinparams we could use
// since everything should stem from that
type asset byte

// AssetCast makes sure that we don't instantiate
func assetCast(attemptedAsset byte) asset {
	switch attemptedAsset {
	case LTCTest:
		return LTCTest
	case BTCTest:
		return BTCTest
	case VTCTest:
		return VTCTest
	}

	return NullAsset
}

func (a *asset) String() string {
	return ByteToAssetString(byte(*a))
}

// generateUniquePairs generates unique asset pairs based on the assets available
func generateUniquePairs(assetList []byte) []Pair {

	assetListLen := len(assetList)
	numPairIndeces := assetListLen * (assetListLen - 1) / 2
	var pairList = make([]Pair, numPairIndeces)
	pairListIndex := 0
	for i, elem := range assetList {
		for lower := i + 1; lower < assetListLen; lower++ {
			pairList[pairListIndex].AssetWant = assetCast(elem)
			pairList[pairListIndex].AssetHave = assetCast(assetList[lower])
			pairListIndex++
		}
	}

	return pairList
}

// GenerateAssetPairs generates unique asset pairs based on the default assets available
func GenerateAssetPairs() []Pair {
	return generateUniquePairs(AssetList())
}

// String is the tostring function for a pair
func (p Pair) String() string {
	return p.AssetWant.String() + "/" + p.AssetHave.String()
}

// making this asset stuff was sort of a mistake, example below
// I just needed an asset struct in the first place

// GetAssociatedCoinParam gets the coinparam parameters related to said asset
func (a asset) GetAssociatedCoinParam() *coinparam.Params {
	switch a {
	case BTCTest:
		return &coinparam.TestNet3Params
	case LTCTest:
		return &coinparam.LiteCoinTestNet4Params
	case VTCTest:
		return &coinparam.VertcoinTestNetParams
	}
	return nil
}
