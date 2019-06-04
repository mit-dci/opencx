package match

// Entry represents either a credit or debit of some asset for some amount
type Entry struct {
	Amount uint64 `json:"amount"`
	Asset  Asset  `json:"asset"`
}

// OrderExecution contains a simple order execution struct. This is what is being used in the clearing
// matching algorithm. We generate order executions so it stays independent of the settlement, and then
// pass those executions upwards.
type OrderExecution struct {
	Pubkey      []byte `json:"pubkey"`
	PrevOrderID []byte `json:"prevorderid"`
	Debited     Entry  `json:"debited"`
	Credited    Entry  `json:"credited"`
}
