package match

import (
	"bytes"
	"encoding/json"
)

// SettlementExecution is the "settlement part" of an execution.
// It defines the operations that should be done to the settlement engine.
type SettlementExecution struct {
	Pubkey   [33]byte `json:"pubkey"`
	Debited  Entry    `json:"debited"`
	Credited Entry    `json:"credited"`
}

// String returns a string representation of the SettlementExecution
func (se *SettlementExecution) String() string {
	// We are ignoring this error because we know the struct is marshallable, since all of the fields are.
	jsonRepresentation, _ := json.Marshal(se)
	return string(jsonRepresentation)
}

// Equal compares one SettlementExecution with another SettlementExecution and returns true if all of the fields are the same.
func (se *SettlementExecution) Equal(otherExec *SettlementExecution) bool {
	if !bytes.Equal(se.Pubkey[:], otherExec.Pubkey[:]) {
		return false
	}
	if se.Debited != otherExec.Debited {
		return false
	}
	if se.Credited != otherExec.Credited {
		return false
	}
	return true
}
