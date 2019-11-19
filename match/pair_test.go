package match

import (
	"crypto/rand"
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
	bothSidesPair = &Pair{
		AssetHave: Asset(0x01),
		AssetWant: Asset(0x01),
	}
)

// TestPairEmptyFirstSideSerialize tests serializing a pair with
// neither side initialized
func TestPairEmptySerialize(t *testing.T) {
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

// TestPairFirstSideSerialize tests serializing a pair with the first
// side initialized
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

// TestPairSecondSideSerialize tests serializing a pair with the
// second side initialized
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

// TestPairBothSideSerialize tests serializing a pair with both sides
// initialized
func TestPairBothSideSerialize(t *testing.T) {
	expectedBuf := [2]byte{0x01, 0x01}
	serializedPair := bothSidesPair.Serialize()
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

// BenchmarkPairSerialize benchmarks the serialization of a Pair
func BenchmarkPairSerialize(b *testing.B) {
	b.StopTimer()
	var err error
	b.SetBytes(2)
	pairToSerialize := &Pair{
		AssetHave: Asset(0x00),
		AssetWant: Asset(0x00),
	}
	randBuf := [2]byte{0x00, 0x00}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Create random bytes
		b.StopTimer()
		if _, err = rand.Read(randBuf[:]); err != nil {
			b.Fatalf("Could not read from random for BenchmarkPairSerialize: %s", err)
			return
		}
		pairToSerialize.AssetHave = Asset(randBuf[0])
		pairToSerialize.AssetHave = Asset(randBuf[1])
		b.StartTimer()
		pairToSerialize.Serialize()
	}
}
