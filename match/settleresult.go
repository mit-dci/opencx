package match

import "encoding/json"

// SettlementResult is the settlement exec and the new balance
type SettlementResult struct {
	NewBal         uint64               `json:"newbal"`
	SuccessfulExec *SettlementExecution `json:"successfulexec"`
}

// String returns the string representation of a settlement result
func (sr *SettlementResult) String() string {
	// this will pass because both are marshallable
	jsonRepresentation, _ := json.Marshal(sr)
	return string(jsonRepresentation)
}
