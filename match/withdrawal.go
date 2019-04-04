package match

import "encoding/binary"

// Withdrawal is a representation of a withdrawal. This is because withdrawals are now signed.
type Withdrawal struct {
	Asset   Asset
	Amount  uint64
	Address string
	// This tells whether or not this is a lightning withdrawal. Default value is false so that makes it easier to not mess up
	Lightning bool
}

// Serialize serializes the withdrawal
func (w *Withdrawal) Serialize() (buf []byte) {
	// bool yes or no
	// Asset [1 byte]
	// Amount [64 bytes]
	// len(address)
	// Address [len(address)]

	var oneorzero byte
	if w.Lightning {
		oneorzero = 0xff
	}
	buf = make([]byte, 66+len(w.Address))
	buf = append(buf, oneorzero)
	buf = append(buf, byte(w.Asset))
	binary.LittleEndian.PutUint64(buf, w.Amount)
	binary.LittleEndian.PutUint64(buf, uint64(len(w.Address)))
	buf = append(buf, []byte(w.Address)...)
	return
}
