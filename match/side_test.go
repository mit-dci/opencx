package match

import "testing"

var (
	buySide               = Buy
	sellSide              = Sell
	expectedBuySerialize  = byte(0x01)
	expectedSellSerialize = byte(0x00)
)

// TestSideStringSell tests the buy side String method
func TestSideStringBuy(t *testing.T) {
	buySideStringRes := buySide.String()
	if buySideStringRes != buyString {
		t.Errorf("Buy Side String implementation incorrect: Expected %s, got %s", buyString, buySideStringRes)
		return
	}
	return
}

// TestSideStringSell tests the sell side String method
func TestSideStringSell(t *testing.T) {
	sellSideStringRes := sellSide.String()
	if sellSideStringRes != sellString {
		t.Errorf("Sell Side String implementation incorrect: Expected %s, got %s", sellString, sellSideStringRes)
		return
	}
	return
}

// TestSideBuyMarshalBinary tests the buy serialization
func TestSideBuyMarshalBinary(t *testing.T) {
	var err error
	var actualBuf []byte
	buy := Buy
	if actualBuf, err = buy.MarshalBinary(); err != nil {
		t.Errorf("Error marshalling buy side binary: %s", err)
		return
	}
	if len(actualBuf) != 1 {
		t.Errorf("Buy side binary marshalling should return an array of length 1. Got %d", len(actualBuf))
		return
	}
	if actualBuf[0] != expectedBuySerialize {
		t.Errorf("Buy side binary marshalling returned incorrect value: Expected %8x, got %8x", expectedBuySerialize, actualBuf)
		return
	}
	return
}

// TestSideSellMarshalBinary tests the sell serialization
func TestSideSellMarshalBinary(t *testing.T) {
	var err error
	var actualBuf []byte
	sell := Sell
	if actualBuf, err = sell.MarshalBinary(); err != nil {
		t.Errorf("Error marshalling sell side binary: %s", err)
		return
	}
	if len(actualBuf) != 1 {
		t.Errorf("Sell side binary marshalling should return an array of length 1. Got %d", len(actualBuf))
		return
	}
	if actualBuf[0] != expectedSellSerialize {
		t.Errorf("Sell side binary marshalling returned incorrect value: Expected %8x, got %8x", expectedSellSerialize, actualBuf)
		return
	}
	return
}

// TestSideBuyUnmarshalBinaryCorrect tests the buy binary
// deserialization for a correct value
func TestSideBuyUnmarshalBinaryCorrect(t *testing.T) {
	var err error
	var emptySide *Side = new(Side)
	if err = emptySide.UnmarshalBinary([]byte{expectedBuySerialize}); err != nil {
		t.Errorf("Error unmarshalling buy side binary: %s", err)
		return
	}

	if *emptySide != Buy {
		t.Errorf("Side unmarshalled was not equal to buy with the correct input %8x", expectedBuySerialize)
		return
	}

	return
}

// TestSideSellUnmarshalBinaryCorrect tests the sell binary
// deserialization for a correct value
func TestSideSellUnmarshalBinaryCorrect(t *testing.T) {
	var err error
	var emptySide Side
	if err = emptySide.UnmarshalBinary([]byte{expectedSellSerialize}); err != nil {
		t.Errorf("Error unmarshalling sell side binary: %s", err)
		return
	}

	if emptySide != Sell {
		t.Errorf("Side unmarshalled was not equal to sell with the correct input %8x", expectedSellSerialize)
		return
	}

	return
}
