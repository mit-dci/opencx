package match

// Pair is a struct that represents a trading pair
type Pair struct {
	AssetWant byte `json:"assetWant"`
	// the AssetHave asset will always be the asset whose balance is checked
	AssetHave byte `json:"assetHave"`
}

// generateUniquePairs generates unique asset pairs based on the assets available
func generateUniquePairs(assetList []byte) []Pair {

	assetListLen := len(assetList)
	numPairIndeces := assetListLen * (assetListLen - 1) / 2
	var pairList = make([]Pair, numPairIndeces)
	pairListIndex := 0
	for i, elem := range assetList {
		for lower := i + 1; lower < assetListLen; lower++ {
			pairList[pairListIndex].AssetWant = elem
			pairList[pairListIndex].AssetHave = assetList[lower]
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
	return ByteToAssetString(p.AssetWant) + "/" + ByteToAssetString(p.AssetHave)
}
