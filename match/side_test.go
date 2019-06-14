package match

import "testing"

var (
	buySide  = Buy
	sellSide = Sell
)

// TestSideStringSell tests the buy side String method
func TestSideStringBuy(t *testing.T) {
	buySideStringRes := buySide.String()
	if buySideStringRes != buyString {
		t.Errorf("Buy Side String implementation incorrect. Expected %s, got %s", buyString, buySideStringRes)
		return
	}
	return
}

// TestSideStringSell tests the sell side String method
func TestSideStringSell(t *testing.T) {
	sellSideStringRes := sellSide.String()
	if sellSideStringRes != sellString {
		t.Errorf("Sell Side String implementation incorrect. Expected %s, got %s", sellString, sellSideStringRes)
		return
	}
	return
}
