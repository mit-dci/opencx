package match

import (
	"crypto/rand"
	"fmt"
	"testing"
)

var (
	emptyAuctionIDExpected = [32]byte{}
	onesAuctionIDExpected  = [32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
)

// TestAuctionIDMarshalBinaryEmpty tests that an empty auction ID has
// the correct length and returns no errors when marshalled to binary
func TestAuctionIDMarshalBinaryEmpty(t *testing.T) {
	var err error
	emptyID := new(AuctionID)
	*emptyID = AuctionID([32]byte{})

	var binaryAnswer []byte
	if binaryAnswer, err = emptyID.MarshalBinary(); err != nil {
		t.Errorf("Marshalling auction ID's should not return errors: %s", err)
		return
	}

	if len(binaryAnswer) != 32 {
		t.Errorf("AuctionID marshalling should return a byte slice of length 32, not %d", len(binaryAnswer))
		return
	}

	var ansArray [32]byte
	copy(ansArray[:], binaryAnswer)
	if ansArray != emptyAuctionIDExpected {
		t.Errorf("Serialized auction ID\n was: %8x\nexpected: %8x\n", ansArray, emptyAuctionIDExpected)
		return
	}
	return
}

// TestAuctionIDMarshalBinaryOnes tests that a non empty auction ID
// has the correct length and returns no errors when marshalled to
// binary
func TestAuctionIDMarshalBinaryOnes(t *testing.T) {
	var err error
	onesID := new(AuctionID)
	*onesID = AuctionID([32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})

	var binaryAnswer []byte
	if binaryAnswer, err = onesID.MarshalBinary(); err != nil {
		t.Errorf("Marshalling auction ID's should not return errors: %s", err)
		return
	}

	if len(binaryAnswer) != 32 {
		t.Errorf("AuctionID marshalling should return a byte slice of length 32, not %d", len(binaryAnswer))
		return
	}

	var ansArray [32]byte
	copy(ansArray[:], binaryAnswer)
	if ansArray != onesAuctionIDExpected {
		t.Errorf("Serialized auction ID\n was: %8x\nexpected: %8x\n", ansArray, onesAuctionIDExpected)
		return
	}
	return
}

// TestAuctionIDUnmarshalBinaryValid tests unmarshalling of a valid
// AuctionID from binary
func TestAuctionIDUnmarshalBinaryValid(t *testing.T) {
	var err error
	onesAuctionID := AuctionID([32]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
	onesBytes := []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ansPtr := new(AuctionID)
	if err = ansPtr.UnmarshalBinary(onesBytes); err != nil {
		t.Errorf("Unmarshalling valid byte array for auctionid should not return an error: %s", err)
		return
	}

	if *ansPtr != onesAuctionID {
		t.Errorf("The AuctionID after unmarshalling %v did not equal the expected value of %v", *ansPtr, onesAuctionID)
		return
	}

	return
}

// TestAuctionIDUnmarshalBinaryShort tests unmarshalling of a short
// AuctionID from binary, checking that an error is actually returned
func TestAuctionIDUnmarshalBinaryShort(t *testing.T) {
	var err error
	shortBytes := []byte{0xde, 0xad, 0xde, 0xad, 0xbe, 0xef, 0xbe, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	ansPtr := new(AuctionID)
	if err = ansPtr.UnmarshalBinary(shortBytes); err == nil {
		t.Errorf("Unmarshalling a short byte array should return an error, instead the error was nil")
		return
	}

	return
}

// TestAuctionIDUnmarshalBinaryLong tests unmarshalling of a long
// AuctionID from binary, checking that an error is actually returned
func TestAuctionIDUnmarshalBinaryLong(t *testing.T) {
	var err error
	longBytes := []byte{0xde, 0xad, 0xde, 0xad, 0xbe, 0xef, 0xbe, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0x00, 0xee, 0x00, 0xee}

	ansPtr := new(AuctionID)
	if err = ansPtr.UnmarshalBinary(longBytes); err == nil {
		t.Errorf("Unmarshalling a long byte array should return an error, instead the error was nil")
		return
	}

	return
}

// TestAuctionIDUnmarshalBinaryInvalid tests unmarshalling of many
// long AuctionIDs from binary, checking that an error is returned in
// all cases
func TestAuctionIDUnmarshalBinaryInvalid(t *testing.T) {
	var err error
	howMany := 40
	ansPtr := new(AuctionID)
	for invalidArr := []byte{}; len(invalidArr) < howMany; invalidArr = append(invalidArr, 0xff) {
		if _, err = rand.Read(invalidArr); err != nil {
			t.Errorf("Error reading from random when creating invalid test array to unmarshal from")
			return
		}
		if len(invalidArr) != 32 {
			t.Run(fmt.Sprintf("UnmarshalLength%d", len(invalidArr)), func(t *testing.T) {
				if err = ansPtr.UnmarshalBinary(invalidArr); err == nil {
					t.Errorf("Unmarshalling an invalid length byte array should return an error, but the error was nil")
					return
				}
			})
		} else {
			t.Run(fmt.Sprintf("UnmarshalLength%d", len(invalidArr)), func(t *testing.T) {
				// it's not an 'invalid' array now
				if err = ansPtr.UnmarshalBinary(invalidArr); err != nil {
					t.Errorf("Unmarshalling a byte array of length 32 should not return an error: %s", err)
					return
				}
			})
		}
	}
	return
}

// BenchmarkAuctionIDMarshalBinary benchmarks the performance of
// marshalling an Auction ID into bytes
func BenchmarkAuctionIDMarshalBinary(b *testing.B) {
	var err error
	b.StopTimer()
	b.ResetTimer()

	idToSerialize := new(AuctionID)
	*idToSerialize = AuctionID([32]byte{0xde, 0xad, 0xde, 0xad, 0xbe, 0xef, 0xbe, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00, 0x00})

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err = idToSerialize.MarshalBinary()
	}
	b.StopTimer()
	if err != nil {
		b.Fatalf("Error when running MarshalBinary on an AuctionID for benchmark: %s", err)
		return
	}
	return
}
