package match

import "encoding/binary"

// Withdrawal is a representation of a withdrawal. This is because withdrawals are now signed.
type Withdrawal struct {
	Asset   Asset
	Amount  uint64
	Address string
	// This tells whether or not this is a lightning withdrawal. Default value is false so that makes it easier
	Lightning bool
}

// Serialize serializes the withdrawal
func (w *Withdrawal) Serialize() (buf []byte) {
	// Asset [1 byte]
	// Amount [64 bytes]
	// len(address)
	// Address [len(address)]

	buf = make([]byte, 65+len(w.Address))
	buf = append(buf, byte(w.Asset))
	binary.LittleEndian.PutUint64(buf, w.Amount)
	binary.LittleEndian.PutUint64(buf, uint64(len(w.Address)))
	buf = append(buf, []byte(w.Address)...)
	return
}
