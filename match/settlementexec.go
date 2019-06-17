package match

import (
	"bytes"
	"encoding/json"
)

// TODO: replace with this once ready
// SettlementExecution is the "settlement part" of an execution.
// It defines the operations that should be done to the settlement engine.
type SettlementExecution struct {
	Pubkey [33]byte `json:"pubkey"`
	Amount uint64   `json:"amount"`
	Asset  Asset    `json:"asset"`
	// SettleType is a type that determines whether or not this is a debit or credit
	Type SettleType `json:"settletype"`
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
	if se.Amount != otherExec.Amount {
		return false
	}
	if se.Asset != otherExec.Asset {
		return false
	}
	if se.Type != otherExec.Type {
		return false
	}
	return true
}
