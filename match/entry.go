package match

import "encoding/json"

// Entry represents either a credit or debit of some asset for some amount
type Entry struct {
	Amount uint64 `json:"amount"`
	Asset  Asset  `json:"asset"`
}

// String returns a json representation of the Entry
func (e *Entry) String() string {
	// we are ignoring this error because we know that the struct is marshallable
	jsonRepresentation, _ := json.Marshal(e)
	return string(jsonRepresentation)
}
