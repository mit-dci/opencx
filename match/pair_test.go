package match

import (
	"testing"
)

var (
	emptyPair     = &Pair{}
	firstSidePair = &Pair{
		AssetWant: Asset(0x01),
	}
	secondSidePair = &Pair{
		AssetHave: Asset(0x01),
	}
)

func TestPairSerialize(t *testing.T) {
	expectedBuf := [2]byte{0x00, 0x00}
	serializedPair := emptyPair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != expectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, expectedBuf)
		return
	}
	return
}

func TestPairFirstSideSerialize(t *testing.T) {
	expectedBuf := [2]byte{0x01, 0x00}
	serializedPair := firstSidePair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != expectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, expectedBuf)
		return
	}
	return

}

func TestPairSecondSideSerialize(t *testing.T) {
	expectedBuf := [2]byte{0x00, 0x01}
	serializedPair := secondSidePair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != expectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, expectedBuf)
		return
	}
	return

}
