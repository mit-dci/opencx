package match

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Side is a representation of buy or sell side
type Side bool

const (
	Buy        = Side(true)  // This should serialize to 0x01
	Sell       = Side(false) // This should serialize to 0x00
	buyString  = "buy"       // just for string representation
	sellString = "sell"      // just for string representation
)

// TODO: Rather than use this string method, implement TextMarshaler
// and TextUnmarshaler

// String returns the string representation of a buy or sell side
func (s Side) String() string {
	if s {
		return buyString
	}
	return sellString
}

// UnmarshalJSON implements the JSON unmarshalling interface
func (s Side) UnmarshalJSON(b []byte) (err error) {
	var str string
	if err = json.Unmarshal(b, &str); err != nil {
		return
	}
	switch strings.ToLower(str) {
	default:
		err = fmt.Errorf("Cannot unmarshal side json, not buy or sell")
		return
	case buyString:
		s = Buy
	case sellString:
		s = Sell
	}
	return
}

// FromString takes a string and, if valid, sets the Side to the
// correct value based on the string
func (s Side) FromString(str string) (err error) {
	switch strings.ToLower(str) {
	default:
		err = fmt.Errorf("Cannot get side from string, not buy or sell")
		return
	case buyString:
		s = Buy
	case sellString:
		s = Sell
	}
	return
}

// MarshalBinary implements the BinaryMarshaler interface from
// encoding.
func (s Side) MarshalBinary() (data []byte, err error) {
	if s {
		data = []byte{0x01}
		return
	}
	data = []byte{0x00}
	return
}

// UnmarshalBinary implements the BinaryUnmarshaler interface from
// encoding.
// This takes a pointer as a receiver because it's not possible nor
// does it make sense to try to modify the value being called.
func (s *Side) UnmarshalBinary(data []byte) (err error) {
	if len(data) != 1 {
		err = fmt.Errorf("Error marshalling binary for a Side, length of data should be 1")
		return
	}
	if data[0] == 0x00 {
		*s = Sell
		return
	} else if data[0] == 0x01 {
		*s = Buy
		return
	}
	err = fmt.Errorf("Error unmarshalling Side, invalid data - should be 0x00 or 0x01")
	return
}
