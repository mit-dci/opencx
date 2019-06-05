package match

// Entry represents either a credit or debit of some asset for some amount
type Entry struct {
	Amount uint64 `json:"amount"`
	Asset  Asset  `json:"asset"`
}

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
*/
type OrderExecution struct {
	OrderID       []byte `json:"orderid"`
	Debited       Entry  `json:"debited"`
	Credited      Entry  `json:"credited"`
	NewAmountWant uint64 `json:"newamtwant"`
	NewAmountHave uint64 `json:"newamthave"`
	Filled        bool   `json:"filled"`
}
