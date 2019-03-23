package match

import (
	"fmt"
	"strings"

	"github.com/mit-dci/lit/coinparam"
)

// Pair is a struct that represents a trading pair
type Pair struct {
	// AssetWant is the asset that buyers want, and that sellers are selling. credit buyers with this.
	AssetWant Asset `json:"assetWant"`
	// AssetHave is the asset that sellers are buying, and that buyers have. credit sellers with this.
	AssetHave Asset `json:"assetHave"`
}

// type aliases are only usually used for codebase refactoring, so make this better when you have time. At some point a struct will probably need to be made.
// like I'm probably going to replace this with just a master list of all the chainhooks / coinparams we could use
// since everything should stem from that

// Asset is a type which represents an asset
type Asset byte

// PrettyString is used to do asset1/asset2 rather than the database-safe asset1_asset2
func (p *Pair) PrettyString() string {
	return p.AssetWant.String() + "/" + p.AssetHave.String()
}

// GenerateAssetPairs generates unique asset pairs based on the coinparams you pass it
func GenerateAssetPairs(coinList []*coinparam.Params) (pairList []*Pair, err error) {

	coinListLen := len(coinList)
	numPairIndeces := coinListLen * (coinListLen - 1) / 2
	pairList = make([]*Pair, numPairIndeces)
	pairListIndex := 0
	for i, elem := range coinList {
		for lower := i + 1; lower < coinListLen; lower++ {
			var assetWant Asset
			if assetWant, err = AssetFromCoinParam(elem); err != nil {
				return
			}

			var assetHave Asset
			if assetHave, err = AssetFromCoinParam(coinList[lower]); err != nil {
				return
			}

			pairList[pairListIndex] = &Pair{
				AssetWant: assetWant,
				AssetHave: assetHave,
			}
			pairListIndex++
		}
	}

	return
}

// Delim is essentially a constant for this struct, I'm sure there are better ways of doing it.
func (p Pair) Delim() string {
	return "_"
}

// String is the tostring function for a pair
func (p Pair) String() string {
	return p.AssetWant.String() + p.Delim() + p.AssetHave.String()
}

// FromString creates a pair object from a string. This is for user input only, hence the slash
func (p *Pair) FromString(pairString string) error {
	strSplit := strings.Split(pairString, "/")

	switch strSplit[0] {
	case BTCTest.String():
		p.AssetWant = BTCTest
	case LTCTest.String():
		p.AssetWant = LTCTest
	case VTCTest.String():
		p.AssetWant = VTCTest
	default:
		return fmt.Errorf("Unsupported Asset")
	}

	switch strSplit[1] {
	case BTCTest.String():
		p.AssetHave = BTCTest
	case LTCTest.String():
		p.AssetHave = LTCTest
	case VTCTest.String():
		p.AssetHave = VTCTest
	default:
		return fmt.Errorf("Unsupported Asset")
	}

	return nil
}

// Serialize serializes the pair into a byte array
func (p Pair) Serialize() []byte {
	return []byte{byte(p.AssetWant), byte(p.AssetHave)}
}

// Deserialize deserializes a byte array into a pair
func (p Pair) Deserialize(buf []byte) (err error) {
	if len(buf) != 2 {
		err = fmt.Errorf("Tried to deserialize ")
	}
	p.AssetWant = Asset(buf[0])
	p.AssetHave = Asset(buf[1])
	return
}
