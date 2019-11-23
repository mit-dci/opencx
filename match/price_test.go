package match

import (
	"testing"
)

// TestPriceDivideByZeroError tests that when trying to return a
// float, a Price with an AmountHave of zero will return an error
func TestPriceDivideByZeroError(t *testing.T) {
	var err error
	divZeroPrice := &Price{
		AmountWant: uint64(12345),
		AmountHave: 0,
	}

	// purposeful err == nil, to make sure that we do in fact have an
	// error when trying to divide by zero
	if _, err = divZeroPrice.ToFloat(); err == nil {
		t.Errorf("Should have an error when dividing by zero, AmountHave should not be zero")
		return
	}
	return
}

// TestPriceDivideByZeroError tests that when trying to return a
// float, a Price with an AmountHave of zero will return a zero value
func TestPriceDivideByZeroValue(t *testing.T) {
	var retPrice float64
	divZeroPrice := &Price{
		AmountWant: uint64(12345),
		AmountHave: 0,
	}

	retPrice, _ = divZeroPrice.ToFloat()

	if retPrice != float64(0) {
		t.Errorf("The price returned by ToFloat should be zero when AmountHave is zero, and there should be an error")
		return
	}
	return
}

// TestPriceDivideByZeroError tests that when trying to return a
// float, a Price with both fields of zero will return an error
func TestPriceZeroZeroError(t *testing.T) {
	var err error
	divZeroPrice := &Price{
		AmountWant: 0,
		AmountHave: 0,
	}

	// purposeful err == nil, to make sure that we do in fact have an
	// error when trying to divide by zero
	if _, err = divZeroPrice.ToFloat(); err == nil {
		t.Errorf("Should have an error when dividing zero by zero, AmountHave should not be zero")
		return
	}
	return
}

// TestPriceDivideByZeroError tests that when trying to return a
// float, a Price with an both fields of zero will return a zero value
func TestPriceZeroZeroValue(t *testing.T) {
	var retPrice float64
	divZeroPrice := &Price{
		AmountWant: 0,
		AmountHave: 0,
	}

	retPrice, _ = divZeroPrice.ToFloat()

	if retPrice != float64(0) {
		t.Errorf("The price returned by ToFloat should be zero when both fields are zero, and there should be an error")
		return
	}
	return
}

// TestPriceCompareInfZero tests that a zero-price order will be less
// than an order with infinite value
func TestPriceCompareInfZero(t *testing.T) {
	infPrice := &Price{
		AmountWant: uint64(1),
		AmountHave: 0,
	}
	zeroPrice := &Price{
		AmountWant: 0,
		AmountHave: uint64(1),
	}

	if leftSide := infPrice.Cmp(zeroPrice); leftSide != 1 {
		t.Errorf("Error, infinite price when compared to zero should be greater")
		return
	}
	if rightSide := zeroPrice.Cmp(infPrice); rightSide != -1 {
		t.Errorf("Error, zero price when compared to infinite should be greater")
		return
	}
	return
}

// TestPriceCompareInfReflexive tests that an infinite price is equal
// to itself
func TestPriceCompareInfReflexive(t *testing.T) {
	infPrice := &Price{
		AmountHave: uint64(1),
		AmountWant: 0,
	}
	secondInfPrice := &Price{
		AmountHave: uint64(2),
		AmountWant: 0,
	}

	if leftSide := infPrice.Cmp(secondInfPrice); leftSide != 0 {
		t.Errorf("Error, two infinite prices should be equal, even if right side is initialized with a greater AmountHave")
		return
	}
	if rightSide := secondInfPrice.Cmp(infPrice); rightSide != 0 {
		t.Errorf("Error, two infinite prices should be equal, even if left side is initialized with a greater AmountHave")
		return
	}
	return
}

// TestPriceCompareInverse tests two prices, one is the inverse of the
// other
func TestPriceCompareInverse(t *testing.T) {
	normalPrice := &Price{
		AmountWant: uint64(505763),
		AmountHave: uint64(37049),
	}
	inversePrice := &Price{
		AmountWant: uint64(37049),
		AmountHave: uint64(505763),
	}

	if leftSide := normalPrice.Cmp(inversePrice); leftSide != 1 {
		t.Errorf("A price that is greater than 1 should not be less than or equal to its inverse")
		return
	}
	if rightSide := inversePrice.Cmp(normalPrice); rightSide != -1 {
		t.Errorf("The inverse of a price greater than 1 should not be greater than or equal to the original price")
		return
	}
	return
}

// TestPriceCompareAlmostInverse tests two prices, one is very close
// to the inverse of the other
func TestPriceCompareAlmostInverse(t *testing.T) {
	normalPrice := &Price{
		AmountWant: uint64(505763),
		AmountHave: uint64(37049),
	}
	almostInversePrice := &Price{
		AmountWant: uint64(37049),
		AmountHave: uint64(505919),
	}

	if leftSide := normalPrice.Cmp(almostInversePrice); leftSide != 1 {
		t.Errorf("Price with AmountHave %d and AmountWant %d should not be greater than or equal to price with AmountHave %d and AmountWant %d", normalPrice.AmountWant, normalPrice.AmountHave, almostInversePrice.AmountWant, almostInversePrice.AmountHave)
		return
	}
	if rightSide := almostInversePrice.Cmp(normalPrice); rightSide != -1 {
		t.Errorf("Price with AmountWant %d and AmountHave %d should not be less than or equal to price with AmountWant %d and AmountHave %d", normalPrice.AmountWant, normalPrice.AmountHave, almostInversePrice.AmountWant, almostInversePrice.AmountHave)
		return
	}
	return
}

// TestPriceFloatValueOne tests that the float value of 1 is actually
// 1
func TestPriceFloatValueOne(t *testing.T) {
	onePrice := &Price{
		AmountWant: 1,
		AmountHave: 1,
	}

	if priceValue, _ := onePrice.ToFloat(); priceValue != float64(1) {
		t.Errorf("A price of one should return 1 when calling ToFloat")
		return
	}
	return
}

// TestPriceFloatErrorOne tests that the error value of a price of 1
// is nil
func TestPriceFloatErrorOne(t *testing.T) {
	onePrice := &Price{
		AmountWant: 1,
		AmountHave: 1,
	}

	if _, err := onePrice.ToFloat(); err != nil {
		t.Errorf("A price of one should not return an error for ToFloat: %s", err)
		return
	}
	return
}

// TestPriceFloatAlmostOne tests the float value of a number very
// close to 1
func TestPriceFloatAlmostOne(t *testing.T) {
	almostOnePrice := &Price{
		AmountWant: uint64(506699),
		AmountHave: uint64(505877),
	}

	var priceValue float64
	var err error
	if priceValue, err = almostOnePrice.ToFloat(); err != nil {
		t.Errorf("A price with AmountWant of %d and AmountHave of %d should not throw an error with ToFloat", almostOnePrice.AmountWant, almostOnePrice.AmountHave)
		return
	}

	// This is pretty arbitrary, it "gets the test to pass" but it is
	// also a reasonable level of tolerance for floating point
	// arithmetic in this case
	floatDifferenceTolerance := float64(0.000000001)

	expectedPrice := float64(1.00162490)
	if priceValue-expectedPrice >= floatDifferenceTolerance {
		t.Errorf("A price with AmountWant of %d and AmountHave of %d should have a price of %f within %f, got %f", almostOnePrice.AmountWant, almostOnePrice.AmountHave, expectedPrice, floatDifferenceTolerance, priceValue)
		return
	}
	return
}
