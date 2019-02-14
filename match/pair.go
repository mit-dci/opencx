package match

import (
	"fmt"
	"strings"

	"github.com/mit-dci/lit/coinparam"
)

// Pair is a struct that represents a trading pair
type Pair struct {
	// AssetWant is the asset that buyers want, and that sellers are selling. credit buyers with this.
	AssetWant asset `json:"assetWant"`
	// AssetHave is the asset that sellers are buying, and that buyers have. credit sellers with this.
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

func (a asset) String() string {
	return ByteToAssetString(byte(a))
}

// generateUniquePairs generates unique asset pairs based on the assets available
func generateUniquePairs(assetList []byte) []*Pair {

	assetListLen := len(assetList)
	numPairIndeces := assetListLen * (assetListLen - 1) / 2
	var pairList = make([]*Pair, numPairIndeces)
	pairListIndex := 0
	for i, elem := range assetList {
		for lower := i + 1; lower < assetListLen; lower++ {
			pairList[pairListIndex] = &Pair{
				AssetWant: assetCast(elem),
				AssetHave: assetCast(assetList[lower]),
			}
			pairListIndex++
		}
	}

	return pairList
}

// GenerateAssetPairs generates unique asset pairs based on the default assets available
func GenerateAssetPairs() []*Pair {
	return generateUniquePairs(AssetList())
}

// Delim is essentially a constant for this struct, I'm sure there are better ways of doing it.
func (p Pair) Delim() string {
	return "_"
}

// String is the tostring function for a pair
func (p Pair) String() string {
	return p.AssetWant.String() + p.Delim() + p.AssetHave.String()
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

// FromString creates a pair object from a string. This is for user input only, hence the slash
func (p *Pair) FromString(pairString string) error {
	strSplit := strings.Split(pairString, "/")

	switch strSplit[0] {
	case assetCast(BTCTest).String():
		p.AssetWant = BTCTest
	case assetCast(LTCTest).String():
		p.AssetWant = LTCTest
	case assetCast(VTCTest).String():
		p.AssetWant = VTCTest
	default:
		return fmt.Errorf("Unsupported Asset")
	}

	switch strSplit[1] {
	case assetCast(BTCTest).String():
		p.AssetHave = BTCTest
	case assetCast(LTCTest).String():
		p.AssetHave = LTCTest
	case assetCast(VTCTest).String():
		p.AssetHave = VTCTest
	default:
		return fmt.Errorf("Unsupported Asset")
	}

	return nil
}
