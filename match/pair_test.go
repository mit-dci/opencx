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
	emptyExpectedBuf      = [2]byte{0x00, 0x00}
	firstSideExpectedBuf  = [2]byte{0x01, 0x00}
	secondSideExpectedBuf = [2]byte{0x00, 0x01}
	bothSidesExpectedBuf  = [2]byte{0x01, 0x01}
)

// TestPairEmptyFirstSideSerialize tests serializing a pair with
// neither side initialized
func TestPairEmptySerialize(t *testing.T) {
	serializedPair := emptyPair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != emptyExpectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, emptyExpectedBuf)
		return
	}
	return
}

// TestPairFirstSideSerialize tests serializing a pair with the first
// side initialized
func TestPairFirstSideSerialize(t *testing.T) {
	serializedPair := firstSidePair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != firstSideExpectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, firstSideExpectedBuf)
		return
	}
	return

}

// TestPairSecondSideSerialize tests serializing a pair with the
// second side initialized
func TestPairSecondSideSerialize(t *testing.T) {
	serializedPair := secondSidePair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != secondSideExpectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, secondSideExpectedBuf)
		return
	}
	return

}

// TestPairBothSideSerialize tests serializing a pair with both sides
// initialized
func TestPairBothSideSerialize(t *testing.T) {
	serializedPair := bothSidesPair.Serialize()
	sPairLen := len(serializedPair)

	if sPairLen != 2 {
		t.Errorf("The length of the serialized pair is %d, it should be 2", len(serializedPair))
		return
	}

	sPairArray := [2]byte{serializedPair[0], serializedPair[1]}
	if sPairArray != bothSidesExpectedBuf {
		t.Errorf("The buffer %8x is not the same as the expected, %8x", sPairArray, bothSidesExpectedBuf)
		return
	}
	return

}

// TestPairEmptyDeserialize tests deserializing a pair for
// neither side initialized
func TestPairEmptyDeserialize(t *testing.T) {
	var err error
	sPair := new(Pair)
	if err = sPair.Deserialize(emptyExpectedBuf[:]); err != nil {
		t.Errorf("Error deserializing empty buffer %8x for pair", emptyExpectedBuf)
		return
	}

	if *sPair != *emptyPair {
		t.Errorf("The empty pair did not deserialize correctly")
		return
	}
	return
}

// TestPairFirstSideDeserialize tests deserializing a pair for the
// first side initialized
func TestPairFirstSideDeserialize(t *testing.T) {
	var err error
	sPair := new(Pair)
	if err = sPair.Deserialize(firstSideExpectedBuf[:]); err != nil {
		t.Errorf("Error deserializing first side buffer %8x for pair", firstSideExpectedBuf)
		return
	}

	if *sPair != *firstSidePair {
		t.Errorf("The first side pair did not deserialize correctly")
		return
	}
	return
}

// TestPairSecondSideDeserialize tests deserializing a pair for the
// second side initialized
func TestPairSecondSideDeserialize(t *testing.T) {
	var err error
	sPair := new(Pair)
	if err = sPair.Deserialize(secondSideExpectedBuf[:]); err != nil {
		t.Errorf("Error deserializing second side buffer %8x for pair", secondSideExpectedBuf)
		return
	}

	if *sPair != *secondSidePair {
		t.Errorf("The second side pair did not deserialize correctly")
		return
	}
	return
}

// TestPairBothSideDeserialize tests deserializing a pair for both
// side initialized
func TestPairBothSideDeserialize(t *testing.T) {
	var err error
	sPair := new(Pair)
	if err = sPair.Deserialize(bothSidesExpectedBuf[:]); err != nil {
		t.Errorf("Error deserializing both sides buffer %8x for pair", bothSidesExpectedBuf)
		return
	}

	if *sPair != *bothSidesPair {
		t.Errorf("The both sides pair did not deserialize correctly")
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

// BenchmarkPairDeserialize benchmarks the serialization of a Pair
func BenchmarkPairDeserialize(b *testing.B) {
	b.StopTimer()
	var err error
	b.SetBytes(2)
	pairToDeserialize := new(Pair)
	pairToDeserialize.AssetWant = Asset(0x00)
	pairToDeserialize.AssetHave = Asset(0x00)
	randBuf := [2]byte{0x00, 0x00}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// Create random bytes
		b.StopTimer()
		if _, err = rand.Read(randBuf[:]); err != nil {
			b.Fatalf("Could not read from random for BenchmarkPairDeserialize: %s", err)
			return
		}
		pairToDeserialize.AssetWant = Asset(0x00)
		pairToDeserialize.AssetHave = Asset(0x00)
		b.StartTimer()
		pairToDeserialize.Deserialize(randBuf[:])
	}
}
