package match

import (
	"bytes"
	"encoding/json"
)

/*
OrderExecution contains a simple order execution struct. This is what is being used in the clearing
matching algorithm. We generate order executions so it stays independent of the settlement, and then
pass those executions upwards. An execution is the output of a matching algorithm, so it's essentially
the change in state from one orderbook matching step to the next.

What won't change when an order is executed:
 - Pubkey
 - AuctionID
 - Nonce
 - Signature
 - TradingPair
 - Side

What definitely will change:
 - AmountWant
 - AmountHave

What can happen?
 - The order is completely filled and should be deleted from the orderbook
In this case the amount debited or credited can be different than the AmountWant and AmountHave specified
due to slippage.
If the order is not deleted then it is partially filled.
In the case that the order is not deleted and it is instead partially filled, the order will have an updated
AmountWant and AmountHave, as well as assets credited and debited.

So what data do we include in an execution?
We don't want to assume anything about the format of the user information, but we can assume that two orders
are somehow distinguishable. This is why we include an order ID.
It's also probably safe to assume that there will be user information associated with an order.
We include:
 - Order identifying information (so an order ID)
 - The amount and asset debited to the user associated with the order ID.
 - The amount and asset credited from the user associated with the order ID.
 - The updated AmountWant
 - The updated AmountHave
 - Whether or not the order was filled completely.
   - If yes, that means the order gets deleted.

The asset for AmountWant and AmountHave can be determined from what they are in the order associated with
the order ID.

Ideally, we want to make sure that either the order is filled completely, or we have the updated
AmountWant and AmountHave. This would be a great place for some sort of enum, but unfortunately
we'll have to go with a bool.

TODO: in the future, when AmountWant and AmountHave are replaced with a new price struct, this will need to
change as well. The NewAmountWant and NewAmountHave can be replaced.

On a typical exchange, say $ per btc, if you place a buy order at a high $/btc and someone else places a sell order
at an even lower $/btc (want/have) after, then your buy order will be executed at your price. However if someone else
places a sell order at a low-ish price, and you place a buy order at a price higher, then it will be executed at a lower price.

Buy orders can only be matched at the price they are placed, or lower.
Sell orders can only be matched at the price they are placed, or higher.
You should never have an order be deleted and it yield
less than you originally requested for the same value you provided.

The good thing is, order executions do not depend on the type of order.
*/
type OrderExecution struct {
	OrderID       []byte `json:"orderid"`
	Debited       Entry  `json:"debited"`
	Credited      Entry  `json:"credited"`
	NewAmountWant uint64 `json:"newamtwant"`
	NewAmountHave uint64 `json:"newamthave"`
	Filled        bool   `json:"filled"`
}

// String returns a json representation of the OrderExecution
func (oe *OrderExecution) String() string {
	// we are ignoring this error because we know that the struct is marshallable. All of the fields are.
	jsonRepresentation, _ := json.Marshal(oe)
	return string(jsonRepresentation)
}

// Equal compares one OrderExecution with another OrderExecution and returns true if all of the fields are the same.
func (oe *OrderExecution) Equal(otherExec *OrderExecution) bool {
	if !bytes.Equal(oe.OrderID, otherExec.OrderID) {
		return false
	}
	if oe.Debited != otherExec.Debited {
		return false
	}
	if oe.Credited != otherExec.Credited {
		return false
	}
	if oe.NewAmountWant != otherExec.NewAmountWant {
		return false
	}
	if oe.NewAmountHave != otherExec.NewAmountHave {
		return false
	}
	if oe.Filled != otherExec.Filled {
		return false
	}
	return true
}
