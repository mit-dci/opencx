package match

import (
	"testing"
)

// TestSOESLenEmpty tests that the buffer has a length of 0 when
// it's empty
func TestSOESLenEmpty(t *testing.T) {
	orderStream := &SmallOrderExecStream{}
	if orderStream.Len() != 0 {
		t.Errorf("Error, empty stream should have a length of zero")
		return
	}
	return
}

// TestSOESIsEmptyActuallyEmpty tests that the buffer returns isempty
// true when it's actually empty
func TestSOESIsEmptyActuallyEmpty(t *testing.T) {
	orderStream := &SmallOrderExecStream{}
	if !orderStream.IsEmpty() {
		t.Errorf("Error, empty stream should return isempty true")
		return
	}
	return
}

// TestSOESLenPushOneFront tests that the buffer returns the correct
// element after a push to the front
func TestSOESPushFrontSingleElement(t *testing.T) {
	var err error
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID: OrderID([32]byte{0xde, 0xad, 0xbe, 0xef}),
		Filled:  true,
	}
	if err = orderStream.PushFront(sampleOrderExec); err != nil {
		t.Errorf("Pushing to an empty stream should not return an error: %s", err)
		return
	}
	var frontOrderExec OrderExecution
	if frontOrderExec, err = orderStream.PeekFront(); err != nil {
		t.Errorf("Error, peekfront should not return an error when we just inserted an element: %s", err)
		return
	}
	if frontOrderExec != sampleOrderExec {
		t.Errorf("Error, order execution returned is not equal to the one inserted")
		return
	}
	return
}

// TestSOESPushBackSingleElement tests that the buffer returns the
// correct element after a push to the back
func TestSOESPushBackSingleElement(t *testing.T) {
	var err error
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID: OrderID([32]byte{0xde, 0xad, 0xbe, 0xef}),
		Filled:  true,
	}
	if err = orderStream.PushBack(sampleOrderExec); err != nil {
		t.Errorf("Pushing to an empty stream should not return an error: %s", err)
		return
	}
	var frontOrderExec OrderExecution
	if frontOrderExec, err = orderStream.PeekBack(); err != nil {
		t.Errorf("Error, peekback should not return an error when we just inserted an element: %s", err)
		return
	}
	if frontOrderExec != sampleOrderExec {
		t.Errorf("Error, order execution returned is not equal to the one inserted")
		return
	}
	return
}

// TestSOESLenOneFront tests that the buffer returns the length after
// a push to the front
func TestSOESLenOneFront(t *testing.T) {
	var err error
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID: OrderID([32]byte{0xde, 0xad, 0xbe, 0xef}),
		Filled:  true,
	}
	if err = orderStream.PushFront(sampleOrderExec); err != nil {
		t.Errorf("Pushing to an empty stream should not return an error: %s", err)
		return
	}
	if orderStream.Len() != 1 {
		t.Errorf("Error, after pushing one element to an empty stream the length is not 1")
		return
	}
	return
}

// TestSOESLenOneBack tests that the buffer returns the length after
// a push to the back
func TestSOESLenOneBack(t *testing.T) {
	var err error
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID: OrderID([32]byte{0xde, 0xad, 0xbe, 0xef}),
		Filled:  true,
	}
	if err = orderStream.PushBack(sampleOrderExec); err != nil {
		t.Errorf("Pushing to an empty stream should not return an error: %s", err)
		return
	}
	if orderStream.Len() != 1 {
		t.Errorf("Error, after pushing one element to an empty stream the length is not 1")
		return
	}
	return
}

// TestSOESIsFull tests that the buffer, after pushed to 256 times,
// returns that it is full
func TestSOESIsFull(t *testing.T) {
	var err error
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID: OrderID([32]byte{0xff, 0xff, 0xde, 0xea, 0xad, 0xdb, 0xbe, 0xee, 0xef, 0xff, 0xff}),
		Filled:  true,
	}
	for i := 0; i < 256; i++ {
		if err = orderStream.PushBack(sampleOrderExec); err != nil {
			t.Errorf("The order exec stream should not error when pushing its %dth element: %s", i, err)
			return
		}
	}
	if !orderStream.IsFull() {
		t.Errorf("The order exec stream should return that it's full when we've inserted 256 elements")
		return
	}
	return
}

// TODO: TestSOESUnderflowPopBack but for Overflow - exact same
// procedure just at the end of the array instead

// TestSOESUnderflowPopBack tests to make sure that the overflow and
// underflow properties are working correctly.
func TestSOESUnderflowPopBack(t *testing.T) {
	var err error
	// This test will add to the front of the buffer, then pop from
	// the back
	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID:       OrderID([32]byte{0xde, 0xad, 0xbe, 0xef}),
		NewAmountHave: 1,
		Filled:        true,
	}

	firstOrder := sampleOrderExec
	firstOrder.NewAmountWant = uint64(9)
	if err = orderStream.PushBack(firstOrder); err != nil {
		t.Errorf("The order exec stream should not error when pushing while non empty: %s", err)
		return
	}

	var orders []OrderExecution
	for i := 0; i < 8; i++ {
		orders = append(orders, sampleOrderExec)
		orders[i].NewAmountWant = uint64(i)
		if err = orderStream.PushFront(orders[i]); err != nil {
			t.Errorf("The order exec stream should not error when pushing an %dth element: %s", i, err)
			return
		}
	}

	//                     0  1       248 249 250 251 252 253 254 255
	// now we should have [9, -, ... ,  7,  6,  5,  4,  3,  2,  1,  0]
	//                     e            s

	// since we remove from the back, the end will have to underflow
	var firstRetOrder OrderExecution
	if firstRetOrder, err = orderStream.PopBack(); err != nil {
		t.Errorf("Should not error when popping from non empty stream: %s", err)
		return
	}

	if firstRetOrder != firstOrder {
		t.Errorf("First order pushed does not equal first order popped: %s", err)
		return
	}

	// So let's pop from the back and make sure that it still works
	var retOrder OrderExecution
	for i := 0; i < 8; i++ {
		if retOrder, err = orderStream.PopBack(); err != nil {
			t.Errorf("Should not error for non-empty stream that has %d elements left: %s", i, err)
			return
		}
		if retOrder.NewAmountWant != uint64(i) {
			t.Errorf("The %dth order pushed was not popped, instead the %dth order was", i, retOrder.NewAmountWant)
			return
		}
		if retOrder != orders[i] {
			t.Errorf("The %dth order in did not equal the %dth order out", i+2, i+2)
			return
		}
	}

	return
}

// BenchmarkSOESPushFront benchmarks the performance of pushing to the
// front in an "inchworm" kind of way
func BenchmarkSOESPushFront(b *testing.B) {
	var err error
	b.StopTimer()
	b.ResetTimer()

	orderStream := &SmallOrderExecStream{}
	sampleOrderExec := OrderExecution{
		OrderID:       OrderID([32]byte{0xf0, 0xde, 0xad, 0xbe, 0xef, 0x0f}),
		NewAmountHave: 191,
		NewAmountWant: 919,
		Filled:        true,
	}

	// procedure: push to the front many times, then repeat pushing
	// to the front and removing from the back
	for i := 0; i < 128; i++ {
		if err = orderStream.PushFront(sampleOrderExec); err != nil {
			b.Fatalf("Should not error when pushing order %d to empty buffer: %s", i, err)
			return
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		err = orderStream.PushFront(sampleOrderExec)
		b.StopTimer()
		if err != nil {
			b.Fatalf("Should not error when pushing to front during loop: %s", err)
			return
		}
		if _, err = orderStream.PopBack(); err != nil {
			b.Fatalf("Should not error when popping from back during loop: %s", err)
			return
		}
		b.StartTimer()
	}

	return
}

// BenchmarkExecChannelPushFront benchmarks the performance of pushing to the
// front of a container/ring in an "inchworm" kind of way
func BenchmarkExecChannelPushFront(b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	orderChannel := make(chan OrderExecution, 256)
	sampleOrderExec := OrderExecution{
		OrderID:       OrderID([32]byte{0xf0, 0xde, 0xad, 0xbe, 0xef, 0x0f}),
		NewAmountHave: 191,
		NewAmountWant: 919,
		Filled:        true,
	}

	// procedure: push to the front many times, then repeat pushing
	// to the front and removing from the back
	for i := 0; i < 128; i++ {
		orderChannel <- sampleOrderExec
	}

	b.StartTimer()
	var dummyOrder OrderExecution
	for i := 0; i < b.N; i++ {
		orderChannel <- sampleOrderExec
		b.StopTimer()
		dummyOrder = <-orderChannel
		b.StartTimer()
	}

	if !dummyOrder.Filled {
		b.Fatalf("Something has gone terribly wrong with this benchmark - Filled is %t", dummyOrder.Filled)
		return
	}

	return
}
