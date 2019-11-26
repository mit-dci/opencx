package match

import "fmt"

// TODO: make this concurrency-safe

// SmallOrderExecStream is a small queue for order execution. It
// implements a circular buffer and has 256 elements max. It is
// intended to hold order executions and be updated quickly.
type SmallOrderExecStream struct {
	execBuf  [256]OrderExecution
	numElems uint16
	start    uint8
	end      uint8
}

// Len returns the length of the order stream queue
func (ss *SmallOrderExecStream) Len() (size int) {
	return int(ss.numElems)
}

// IsEmpty returns whether or not the queue is empty
func (ss *SmallOrderExecStream) IsEmpty() (empty bool) {
	return ss.numElems == 0
}

// IsFull returns whether or not the queue is full
func (ss *SmallOrderExecStream) IsFull() (full bool) {
	return ss.numElems == 256
}

// PushFront inserts an element at the front of the queue, the
// function errors if there is not enough space
func (ss *SmallOrderExecStream) PushFront(orderExec OrderExecution) (err error) {
	if ss.IsFull() {
		err = fmt.Errorf("Error, order exec queue is full, cannot push to front")
		return
	}
	// NOTE: this code takes ADVANTAGE of the fact that underflow is a
	// thing with unsigned integers.
	// If the start is 0 and we need to add something to the end of
	// the array, then we will just underflow the index!
	// We will never get an out of bounds error.

	if ss.numElems == 0 {
		ss.start = 255
		ss.end = 255
		ss.numElems = 1
		ss.execBuf[255] = orderExec
		return
	}
	ss.numElems++

	ss.start--
	ss.execBuf[ss.start] = orderExec
	return
}

// PushBack inserts an element at the back of the queue, the
// function errors if there is not enough space
func (ss *SmallOrderExecStream) PushBack(orderExec OrderExecution) (err error) {
	if ss.IsFull() {
		err = fmt.Errorf("Error, order exec queue is full, cannot push to back")
		return
	}
	// NOTE: this code takes ADVANTAGE of the fact that overflow is a
	// thing with unsigned integers.
	// If the end is 255 and we need to add something to the beginning of
	// the array, then we will just overflow the index!
	// We will never get an out of bounds error.

	if ss.numElems == 0 {
		ss.start = 0
		ss.end = 0
		ss.numElems = 1
		ss.execBuf[0] = orderExec
		return
	}
	ss.numElems++

	ss.end++
	ss.execBuf[ss.end] = orderExec
	return
}

// PopFront returns an element at the front of the queue and removes
// it from the front, the function errors if the buffer is empty.
func (ss *SmallOrderExecStream) PopFront() (orderExec OrderExecution, err error) {
	if ss.IsEmpty() {
		err = fmt.Errorf("Error, order exec queue is empty, nothing to pop from front")
		return
	}
	ss.numElems--

	// NOTE: this code takes ADVANTAGE of the fact that overflow is a
	// thing with unsigned integers - see note in PushBack and
	// PushFront for how this works.
	orderExec = ss.execBuf[ss.start]
	ss.start++
	return
}

// PopBack returns an element at the back of the queue and removes
// it from the back, the function errors if the buffer is empty.
func (ss *SmallOrderExecStream) PopBack() (orderExec OrderExecution, err error) {
	if ss.IsEmpty() {
		err = fmt.Errorf("Error, order exec queue is empty, nothing to pop from back")
		return
	}
	ss.numElems--

	// NOTE: this code takes ADVANTAGE of the fact that overflow is a
	// thing with unsigned integers - see note in PushBack and
	// PushFront for how this works.
	orderExec = ss.execBuf[ss.end]
	ss.end--
	return
}

// PeekFront returns an element at the front of the queue, the
// function errors if the buffer is empty.
func (ss *SmallOrderExecStream) PeekFront() (orderExec OrderExecution, err error) {
	if ss.IsEmpty() {
		err = fmt.Errorf("Error, order exec queue is empty, nothing to peek at from front")
		return
	}
	orderExec = ss.execBuf[ss.start]
	return
}

// PeekBack returns an element at the back of the queue, the
// function errors if the buffer is empty.
func (ss *SmallOrderExecStream) PeekBack() (orderExec OrderExecution, err error) {
	if ss.IsEmpty() {
		err = fmt.Errorf("Error, order exec queue is empty, nothing to peek at from back")
		return
	}
	orderExec = ss.execBuf[ss.end]
	return
}

// Flush removes all elements from the circular buffer, returning it
// in a slice, erroring if the buffer is empty
func (ss *SmallOrderExecStream) Flush() (orderExecs []OrderExecution, err error) {
	if ss.IsEmpty() {
		err = fmt.Errorf("Error, order exec queue is empty, nothing to flush")
		return
	}
	orderExecs = make([]OrderExecution, ss.numElems)
	if ss.start < ss.end {
		ss.numElems = 0
		copy(orderExecs, ss.execBuf[ss.start:ss.end])
		return
	}
	if ss.start == ss.end {
		ss.numElems = 0
		orderExecs[0] = ss.execBuf[ss.start]
		return
	}
	startExclusive := 256 - int16(ss.start)
	copy(orderExecs[0:startExclusive], ss.execBuf[ss.start:])
	copy(orderExecs[startExclusive:], ss.execBuf[:ss.end+1])
	ss.numElems = 0
	return
}
