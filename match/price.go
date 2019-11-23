package match

import (
	"fmt"
	"math/big"
)

// Price represents an exchange rate. It's basically a fancy fraction. It follows the Want / Have method of doing things.
// TODO: The price and side can be represented as a single struct that contains the Asset the user wants (AssetWant), the Asset that the user has (AssetHave), the amount of AssetHave the user is willing to give up, and the amount of AssetWant that the user would like.
// This way there is no reason to have a "side" field, and prices can be represented more fairly, where you do not need to worry about precision other than the maximum precision of the asset, which is usually within a uint64.
// We don't want this to be a big int because that means it can't really be sent over the wire. We're not multiple precision here, but we do want some
// standard, reasonable level of precision
type Price struct {
	AmountWant uint64
	AmountHave uint64
}

// Note on the Want / Have model: It makes sense from an exchange perspective, but in reality "side", "price", and "volume" are all connected.

// ToFloat converts the price to a float value
func (p *Price) ToFloat() (price float64, err error) {
	if p.AmountHave == 0 {
		err = fmt.Errorf("AmountHave cannot be 0 to convert to float")
		return
	}
	price = float64(p.AmountWant) / float64(p.AmountHave)
	return
}

// Cmp compares p and otherPrice and returns:
//
//   -1 if x <  y
//    0 if x == y (incl. -0 == 0, -Inf == -Inf, and +Inf == +Inf)
//   +1 if x >  y
//
func (p *Price) Cmp(otherPrice *Price) (compIndicator int) {
	// If we want to compare a/b and c/d, then we can just compare a*d
	// and b*c
	numeratorOne := new(big.Int).SetUint64(p.AmountWant)
	numeratorTwo := new(big.Int).SetUint64(otherPrice.AmountWant)
	denominatorOne := new(big.Int).SetUint64(p.AmountHave)
	denominatorTwo := new(big.Int).SetUint64(otherPrice.AmountHave)

	// because this actually sets, we don't need to assign
	numeratorOne.Mul(numeratorOne, denominatorTwo)
	numeratorTwo.Mul(numeratorTwo, denominatorOne)

	// a/b ?= c/d <=> ad ?= bc
	compIndicator = numeratorOne.Cmp(numeratorTwo)
	return
}
