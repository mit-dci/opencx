package match

const (
	// NullAsset is a constant used to represent an unsupported or undefined asset
	NullAsset = 0x00
	// BTCTest is a constant used to represent a BTC Test net token
	BTCTest = 0x01
	// VTCTest is a constant used to represent a VTC Test net token
	VTCTest = 0x02
	// LTCTest is a constant used to represent a LTC Test net token
	LTCTest = 0x03
)

// AssetList returns the list of assets that the exchange supports
func AssetList() []byte {
	return []byte{BTCTest, VTCTest, LTCTest}
}

// largeAssetList is something used for testing the generateassetpairs function, this should be put into a unit test once tests are written
func largeAssetList() []byte {
	return []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}
}

// ByteToAssetString turns a byte (which would be in a "pair") into a string
func ByteToAssetString(assetByte byte) string {
	switch assetByte {
	case BTCTest:
		return "BTCTest"
	case VTCTest:
		return "VTCTest"
	case LTCTest:
		return "LTCTest"
	default:
		return "Unsupported asset"
	}
}
