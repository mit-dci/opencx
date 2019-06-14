package match

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// TODO: Order, Side, Price, User abstraction: The Price should really be the pair {amountHave,amountWant}, and we should be comparing Prices by doing fraction comparison.
// TODO: Create arithmetic for orders, work out decimals, make testable.

// LimitOrder represents a limit order, implementing the order interface
type LimitOrder struct {
	Pubkey      [33]byte `json:"pubkey"`
	Side        Side     `json:"side"`
	TradingPair Pair     `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amounthave"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64 `json:"amountwant"`
}

// Price gets a float price for the order. This determines how it will get matched. The exchange should figure out if it can take some of the
func (l *LimitOrder) Price() (price float64, err error) {
	if l.AmountWant == 0 || l.AmountHave == 0 {
		err = fmt.Errorf("Cannot calculate price if AmountWant or AmountHave is 0")
		return
	}
	price = float64(l.AmountWant) / float64(l.AmountHave)
	return
}

// Serialize serializes an order, possible replay attacks here since this is what you're signing?
func (l *LimitOrder) Serialize() (buf []byte, err error) {
	intermediate := new(bytes.Buffer)
	if err = binary.Write(intermediate, binary.LittleEndian, *l); err != nil {
		err = fmt.Errorf("Error writing limit order to binary for serialize: %s", err)
		return
	}
	if _, err = intermediate.Read(buf); err != nil {
		err = fmt.Errorf("Error reading from intermediate into buffer for serialize: %s", err)
		return
	}
	return
}
