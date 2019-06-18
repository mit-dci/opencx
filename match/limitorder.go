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

// GenerateOrderFill creates an execution that will fill an order (AmountHave at the end is 0) and provides an order and settlement execution.
// This does not assume anything about the price of the order, as we can't infer what price the order was
// placed at.
// TODO: Figure out whether or not these should be pointers
func (l *LimitOrder) GenerateOrderFill(orderID *OrderID, execPrice float64) (orderExec OrderExecution, setExecs []*SettlementExecution, err error) {

	if l.AmountHave == 0 {
		err = fmt.Errorf("Error generating order fill: empty order, the AmountHave cannot be 0")
		return
	}

	if execPrice == float64(0) {
		err = fmt.Errorf("Error generating order fill: price cannot be zero")
		return
	}
	// So we want to make sure that we're filling this, so first calculate the amountWantToFill and make sure it equals amountHave
	var amountToDebit uint64
	var debitAsset Asset
	var creditAsset Asset
	if l.Side == Buy {
		debitAsset = l.TradingPair.AssetWant
		creditAsset = l.TradingPair.AssetHave
	} else if l.Side == Sell {
		debitAsset = l.TradingPair.AssetHave
		creditAsset = l.TradingPair.AssetWant
	} else {
		err = fmt.Errorf("Error generating order fill from price, order is not buy or sell side, it's %s side", l.Side)
		return
	}
	amountToDebit = uint64(float64(l.AmountHave) * execPrice)

	// IMPORTANT! These lines:
	// > OrderID: make([]byte, len(orderID),
	// > copy(execution.OrderID, orderID)
	// are EXTREMELY IMPORTANT because there's pretty much no other way to copy
	// bytes. This was a really annoying issue to debug
	// Finally generate execution
	orderExec = OrderExecution{
		OrderID:       *orderID,
		NewAmountWant: 0,
		NewAmountHave: 0,
		Filled:        true,
	}
	debitSetExec := SettlementExecution{
		Amount: amountToDebit,
		Asset:  debitAsset,
		Type:   Debit,
	}
	creditSetExec := SettlementExecution{
		Amount: l.AmountHave,
		Asset:  creditAsset,
		Type:   Credit,
	}

	copy(debitSetExec.Pubkey[:], l.Pubkey[:])
	copy(creditSetExec.Pubkey[:], l.Pubkey[:])

	setExecs = append(setExecs, &debitSetExec)
	setExecs = append(setExecs, &creditSetExec)
	return
}

// GenerateExecutionFromPrice generates a trade execution from a price and an amount to fill. This is intended to be
// used by the matching engine when a price is determined for this order to execute at.
// amountToFill refers to the amount of AssetWant that can be filled. So the other side's "AmountHave" can be passed
// in as a parameter. The order ID will be filled in, as it's being passed as a parameter.
// This returns a fillRemainder, which is the amount that is left over from amountToFill after
// filling orderID at execPrice and amountToFill
func (l *LimitOrder) GenerateExecutionFromPrice(orderID *OrderID, execPrice float64, amountToFill uint64) (orderExec OrderExecution, setExecs []*SettlementExecution, fillRemainder uint64, err error) {
	// If it's a buy side, AmountWant is assetWant, and AmountHave is assetHave - but price is something different, price is want/have.
	// So to convert from amountWant (amountToFill) to amountHave we need to multiple amountToFill by 1/execPrice
	var amountWantToFill uint64
	var debitAsset Asset
	var creditAsset Asset
	if l.Side == Buy {
		debitAsset = l.TradingPair.AssetHave
		creditAsset = l.TradingPair.AssetWant
	} else if l.Side == Sell {
		debitAsset = l.TradingPair.AssetWant
		creditAsset = l.TradingPair.AssetHave
	} else {
		err = fmt.Errorf("Error generating execution from price, order is not buy or sell side, it's %s side", l.Side)
		return
	}
	amountWantToFill = uint64(float64(amountToFill) * execPrice)

	// Now that we have this value, we'll generate the execution.
	// What should we do if the amountWantToFill is greater than the amountHave?
	// We know that if it's equal we'll just mark the order as filled, return the debited, return the credited, and move on.
	// In what cases would it be greater?
	// Well we don't want to credit them more than they have. So we can only fill it up to a certain amount.
	if amountWantToFill >= l.AmountHave {
		fillRemainder = amountWantToFill - l.AmountHave
		if orderExec, setExecs, err = l.GenerateOrderFill(orderID, execPrice); err != nil {
			err = fmt.Errorf("Error generating order fill while generating exec for price: %s", err)
			return
		}
	} else {
		// TODO: Check if this len(orderID) as size is correct
		orderExec = OrderExecution{
			OrderID:       *orderID,
			NewAmountWant: l.AmountWant - amountToFill,
			NewAmountHave: l.AmountHave - amountWantToFill,
			Filled:        false,
		}

		debitSetExec := SettlementExecution{
			Amount: amountToFill,
			Asset:  debitAsset,
			Type:   Debit,
		}
		creditSetExec := SettlementExecution{
			Amount: amountWantToFill,
			Asset:  creditAsset,
			Type:   Credit,
		}

		copy(debitSetExec.Pubkey[:], l.Pubkey[:])
		copy(creditSetExec.Pubkey[:], l.Pubkey[:])

		setExecs = append(setExecs, &debitSetExec)
		setExecs = append(setExecs, &creditSetExec)
	}

	return
}
