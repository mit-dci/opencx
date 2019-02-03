package match

import (
	"fmt"
)

// RequestForQuoteOrder represents an order of type RFQ implementing the order interface
type RequestForQuoteOrder struct {
	Client      string `json:"username"`
	TradingPair Pair   `json:"pair"`
	AmountHave  uint64 `json:"amount"`
}

// Type returns the type of order. It's in the interface because every type of order needs one
func (r *RequestForQuoteOrder) Type() string {
	return "RFQ"
}

// Price returns the price of the order. For this it's a request for quote, price isn't known for a while so we return an error for now.
// If we include a channel that actually requests to a server for the quote and returns the price then this will take that into account
// somehow.
func (r *RequestForQuoteOrder) Price() (float64, error) {
	return 0, fmt.Errorf("Request for quote price not implemented for order interface")
}
