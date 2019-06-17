package match

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SettleType is a type that represents either a credit or a debit
type SettleType bool

const (
	Debit = SettleType(iota%2 == 0)
	Credit
)
const (
	debitString  = "debit"
	creditString = "credit"
)

// String returns the string representation of a credit or debit
func (st *SettleType) String() string {
	if *st {
		return debitString
	}
	return creditString
}

func (st *SettleType) UnmarshalJSON(b []byte) (err error) {
	var str string
	if err = json.Unmarshal(b, &str); err != nil {
		return
	}
	switch strings.ToLower(str) {
	default:
		err = fmt.Errorf("Cannot unmarshal settletype json, not credit or debit")
		return
	case debitString:
		*st = Debit
	case creditString:
		*st = Credit
	}

	return
}
