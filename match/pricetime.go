package match

import (
	"fmt"
)

// MatchPTPAlgorithm runs matching on an orderbook that is unsorted or unprioritized.
// These get sorted then matched efficiently.
// TODO: Implement PrioritizeOrderbookPTP and then uncomment this method, write tests
// func MatchPTPAlgorithm(book map[float64][]*LimitOrderIDPair) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error) {

// 	var buyOrders []*LimitOrderIDPair
// 	var sellOrders []*LimitOrderIDPair
// 	if buyOrders, sellOrders, err = PrioritizeOrderbookPTP(book); err != nil {
// 		err = fmt.Errorf("Error prioritizing orders for MatchPTPAlgorithm: %s", err)
// 		return
// 	}

// 	if orderExecs, settlementExecs, err = MatchPrioritizedOrders(buyOrders, sellOrders); err != nil {
// 		err = fmt.Errorf("Error matching prioritized orders for MatchPTPAlgorithm: %s", err)
// 		return
// 	}

// 	return
// }

// MatchPrioritizedOrders matches separated buy and sell orders that are properly sorted in price-time priority.
// These are the orders that should match.
// This should never return a list of order executions containing the same ID for more than one execution
func MatchPrioritizedOrders(buyOrders []*LimitOrderIDPair, sellOrders []*LimitOrderIDPair) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error) {
	// Lists should be in priority order starting at 0
	for len(buyOrders) > 0 && len(sellOrders) > 0 && buyOrders[0].Price <= sellOrders[0].Price {
		// Ahh whatever we can be a little inefficient space-wise, just add em all to the list
		// and optimize later

		// // If the order ID is different then we're done with this order execution
		// if currBuyOrderExec.OrderID != *currBuy.OrderID {
		// 	orderExecs = append(orderExecs, currBuyOrderExec)
		// 	currBuyOrderExec = new(OrderExecution)
		// }
		// if currSellOrderExec.OrderID != *currSell.OrderID {
		// 	orderExecs = append(orderExecs, currSellOrderExec)
		// 	currSellOrderExec = new(OrderExecution)
		// }

		// // If the order pubkey is different then we're done with this settlement execution
		// if currBuySetExec.Pubkey != currBuy.Order.Pubkey {
		// 	settlementExecs = append(settlementExecs, currBuySetExec)
		// 	currBuySetExec = new(SettlementExecution)
		// }
		// if currSellSetExec.Pubkey != currSell.Order.Pubkey {
		// 	settlementExecs = append(settlementExecs, currSellSetExec)
		// 	currSellSetExec = new(SettlementExecution)
		// }

		// If sell was first, use that price
		var prSellExec OrderExecution
		var prBuyExec OrderExecution
		var prelimSettlementExecs []*SettlementExecution
		if prBuyExec, prSellExec, prelimSettlementExecs, err = MatchTwoOpposite(buyOrders[0], sellOrders[0]); err != nil {
			err = fmt.Errorf("Error matching orders")
		}

		// Set new amounts because we either want final amounts (when loop conds won't satisfy)
		// or we want a fill
		buyOrders[0].Order.AmountHave = prBuyExec.NewAmountHave
		buyOrders[0].Order.AmountWant = prBuyExec.NewAmountWant

		sellOrders[0].Order.AmountHave = prSellExec.NewAmountHave
		sellOrders[0].Order.AmountWant = prSellExec.NewAmountWant

		if len(buyOrders) == 0 || len(sellOrders) == 0 || buyOrders[0].Price > sellOrders[0].Price {
			// If we will be done, make sure to add the result
			orderExecs = append(orderExecs, []*OrderExecution{&prBuyExec, &prSellExec}...)
		} else {
			// If we're gonna keep going, then only add if we're filled.
			if prSellExec.Filled {
				sellOrders = sellOrders[1:]
				orderExecs = append(orderExecs, &prSellExec)
			}
			if prBuyExec.Filled {
				buyOrders = buyOrders[1:]
				orderExecs = append(orderExecs, &prBuyExec)
			}
		}

		// we keep all of the settlements no matter what because the rates may be
		// changing (due to time priority)
		settlementExecs = append(settlementExecs, prelimSettlementExecs...)
	}
	return
}

// PrioritizeOrderbookPTP prioritizes orders in a map representation of an orderbook by price-time priority.
// It then separates that into buy and sell lists, which get returned.
// This makes it easy to put in to the MatchPrioritizedOrders algorithm.
// TODO: Implement this. It's not really necessary but helpful
// func PrioritizeOrderbookPTP(book map[float64][]*LimitOrderIDPair) (buyOrders []*LimitOrderIDPair, sellOrders []*LimitOrderIDPair, err error) {
// 	logging.Fatalf("UNIMPLEMENTED!!!")
// 	return
// }

// MatchTwo matches a buy order order with the sell order supplied as an argument, giving this order priority.
func MatchTwoOpposite(buyLp *LimitOrderIDPair, sellLp *LimitOrderIDPair) (buyExec OrderExecution, sellExec OrderExecution, settlementExecs []*SettlementExecution, err error) {

	if buyLp.Order.Side != Buy || sellLp.Order.Side != Sell {
		err = fmt.Errorf("Invalid input, buy LimitOrderIDPair was not buy or sell LimitOrderIDPair was not sell")
		return
	}
	buyExec = OrderExecution{
		OrderID: *buyLp.OrderID,
	}
	sellExec = OrderExecution{
		OrderID: *sellLp.OrderID,
	}

	// So we have two LimitOrderIDPairs, and we're going to be generating executions for each
	// as well as settlement executions, because we'll be checking price/time
	if buyLp.Timestamp.UnixNano() > sellLp.Timestamp.UnixNano() {
		// if the sell order amount have >= buy order amount want then we can just try to fill the buy
		// order at the sell order's price
		if sellLp.Order.AmountHave >= buyLp.Order.AmountWant {
			var sellSetExecs []*SettlementExecution
			if sellExec, sellSetExecs, _, err = sellLp.Order.GenerateExecutionFromPrice(sellLp.OrderID, sellLp.Price, buyLp.Order.AmountHave); err != nil {
				err = fmt.Errorf("Error generating exec from price for sell MatchTwoOpposite: %s", err)
				return
			}
			// so we have this sell order exec, now we fill the buy order at the sell price
			var buySetExecs []*SettlementExecution
			if buyExec, buySetExecs, err = buyLp.Order.GenerateOrderFill(buyLp.OrderID, sellLp.Price); err != nil {
				err = fmt.Errorf("Error generating fill from price for buy MatchTwoOpposite: %s", err)
				return
			}

			// append to the settlement execs, we have all we need, let's exit
			settlementExecs = append(settlementExecs, sellSetExecs...)
			settlementExecs = append(settlementExecs, buySetExecs...)

			// We are done. Quit early
			return
		}

		// otherwise the sell order will be filled by the buy order, so we generate a fill for the sell order.
		// The sellOrder.AmountHave < buyOrder.AmountWant, so we will generate a fill for the sell order
		var sellSetExecs []*SettlementExecution
		if sellExec, sellSetExecs, err = sellLp.Order.GenerateOrderFill(sellLp.OrderID, sellLp.Price); err != nil {
			err = fmt.Errorf("Error generating fill for sell order for MatchTwoOpposite: %s", err)
			return
		}

		var buySetExecs []*SettlementExecution
		if buyExec, buySetExecs, _, err = buyLp.Order.GenerateExecutionFromPrice(buyLp.OrderID, sellLp.Price, sellLp.Order.AmountHave); err != nil {
			err = fmt.Errorf("Error generating buy exec from price for MatchTwoOpposite: %s", err)
			return
		}

		// append to the settlement execs, we have all we need, let's exit
		settlementExecs = append(settlementExecs, sellSetExecs...)
		settlementExecs = append(settlementExecs, buySetExecs...)

		// We are done. Quit early
		return
	}

	if buyLp.Order.AmountHave >= sellLp.Order.AmountWant {
		// The sell order can be filled so we generate a fill for that and a normal, leftover exec for the buy order
		var sellSetExecs []*SettlementExecution
		if sellExec, sellSetExecs, err = sellLp.Order.GenerateOrderFill(sellLp.OrderID, buyLp.Price); err != nil {
			err = fmt.Errorf("Error generating fill for sell order for MatchTwoOpposite: %s", err)
			return
		}

		var buySetExecs []*SettlementExecution
		if buyExec, buySetExecs, _, err = buyLp.Order.GenerateExecutionFromPrice(buyLp.OrderID, buyLp.Price, sellLp.Order.AmountHave); err != nil {
			err = fmt.Errorf("Error generating buy exec from price for MatchTwoOpposite: %s", err)
			return
		}

		// append to the settlement execs, we have all we need, let's exit
		settlementExecs = append(settlementExecs, sellSetExecs...)
		settlementExecs = append(settlementExecs, buySetExecs...)

		// We are done. Quit early
		return
	}
	// The sell order > buy order so we generate partial fill for sell and full for buy
	var sellSetExecs []*SettlementExecution
	if sellExec, sellSetExecs, _, err = sellLp.Order.GenerateExecutionFromPrice(sellLp.OrderID, buyLp.Price, buyLp.Order.AmountHave); err != nil {
		err = fmt.Errorf("Error generating exec from price for sell MatchTwoOpposite: %s", err)
		return
	}
	// so we have this sell order exec, now we fill the buy order at the sell price
	var buySetExecs []*SettlementExecution
	if buyExec, buySetExecs, err = buyLp.Order.GenerateOrderFill(buyLp.OrderID, buyLp.Price); err != nil {
		err = fmt.Errorf("Error generating fill from price for buy MatchTwoOpposite: %s", err)
		return
	}

	// append to the settlement execs, we have all we need, let's exit
	settlementExecs = append(settlementExecs, sellSetExecs...)
	settlementExecs = append(settlementExecs, buySetExecs...)

	// We are done. Quit early
	return
}
