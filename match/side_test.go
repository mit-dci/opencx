package match

import "testing"

var (
	buySide               = Buy
	sellSide              = Sell
	expectedBuySerialize  = byte(0x01)
	expectedSellSerialize = byte(0x00)
	incorrectArrays       = [20][1]byte{
		[1]byte{0x0f},
		[1]byte{0x04},
		[1]byte{0xbb},
		[1]byte{0xd0},
		[1]byte{0x11},
		[1]byte{0xe0},
		[1]byte{0xde},
		[1]byte{0xf0},
		[1]byte{0xb4},
		[1]byte{0xab},
		[1]byte{0x12},
		[1]byte{0xff},
		[1]byte{0xbe},
		[1]byte{0x13},
		[1]byte{0xa4},
		[1]byte{0xab},
		[1]byte{0xfe},
		[1]byte{0x0e},
		[1]byte{0x10},
		[1]byte{0x20},
	}
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

// TestSideUnmarshalBinaryWrongSize2 tests the side binary
// deserialization for an array of size 2
func TestSideUnmarshalBinaryWrongSize2(t *testing.T) {
	var err error
	var emptySide Side
	if err = emptySide.UnmarshalBinary([]byte{expectedSellSerialize, 0xff}); err == nil {
		t.Errorf("UnmarshalBinary should return an error when a byte array of size 2(>1) is passed as input")
		return
	}

	return
}

// TestSideUnmarshalBinaryWrongSize3 tests the side binary
// deserialization for an array of size 3
func TestSideUnmarshalBinaryWrongSize3(t *testing.T) {
	var err error
	var emptySide Side
	if err = emptySide.UnmarshalBinary([]byte{expectedSellSerialize, 0xff, 0xfe}); err == nil {
		t.Errorf("UnmarshalBinary should return an error when a byte array of size 3(>1) is passed as input")
		return
	}

	return
}

// TestSideUnmarshalBinaryWrongSize4 tests the side binary
// deserialization for an array of size 4
func TestSideUnmarshalBinaryWrongSize4(t *testing.T) {
	var err error
	var emptySide Side
	if err = emptySide.UnmarshalBinary([]byte{0x00, expectedSellSerialize, 0xff, 0x01}); err == nil {
		t.Errorf("UnmarshalBinary should return an error when a byte array of size 4 (>1) is passed as input")
		return
	}

	return
}

// TestSideUnmarshalBinaryWrongValue tests the side binary
// deserialization for an array of one element that has an incorrect
// value (not 0x00 or 0x01)
func TestSideUnmarshalBinaryWrongValue(t *testing.T) {
	var err error
	var emptySide Side
	if err = emptySide.UnmarshalBinary([]byte{0x02}); err == nil {
		t.Errorf("UnmarshalBinary should return an error when a byte array of one element that isn't 0x00 or 0x01 is passed as input")
		return
	}

	return
}

// TestSideUnmarshalBinaryManyWrongValue tests the side binary
// deserialization for many incorrect values
func TestSideUnmarshalBinaryManyWrongValue(t *testing.T) {
	var err error
	var emptySide Side
	for _, elem := range incorrectArrays {
		if err = emptySide.UnmarshalBinary(elem[:]); err == nil {
			t.Errorf("UnmarshalBinary should return an error for input %8x which is not 0x00 or 0x01", elem[0])
			return
		}
	}

	return
}
