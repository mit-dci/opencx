package match

import (
	"fmt"

	"github.com/mit-dci/opencx/logging"
)

// MatchPTPAlgorithm runs matching on an orderbook that is unsorted or unprioritized.
// These get sorted then matched efficiently.
func MatchPTPAlgorithm(book map[float64][]*LimitOrderIDPair) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error) {

	var buyOrders []*LimitOrderIDPair
	var sellOrders []*LimitOrderIDPair
	if buyOrders, sellOrders, err = PrioritizeOrderbookPTP(book); err != nil {
		err = fmt.Errorf("Error prioritizing orders for MatchPTPAlgorithm: %s", err)
		return
	}

	if orderExecs, settlementExecs, err = MatchPrioritizedOrders(buyOrders, sellOrders); err != nil {
		err = fmt.Errorf("Error matching prioritized orders for MatchPTPAlgorithm: %s", err)
		return
	}

	return
}

// MatchPrioritizedOrders matches separated buy and sell orders that are properly sorted in price-time priority.
// These are the orders that should match.
func MatchPrioritizedOrders(buyOrders []*LimitOrderIDPair, sellOrders []*LimitOrderIDPair) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error) {
	// Lists should be in priority order starting at 0
	var currBuyOrderExec *OrderExecution
	var currSellOrderExec *OrderExecution
	var currBuySetExec *SettlementExecution
	var currSellSetExec *SettlementExecution
	for len(buyOrders) > 0 && len(sellOrders) > 0 && buyOrders[0].Price <= sellOrders[0].Price {
		currBuy := buyOrders[0]
		currSell := sellOrders[0]

		// If the order ID is different then we're done with this order execution
		if currBuyOrderExec.OrderID != *currBuy.OrderID {
			orderExecs = append(orderExecs, currBuyOrderExec)
			currBuyOrderExec = new(OrderExecution)
		}
		if currSellOrderExec.OrderID != *currSell.OrderID {
			orderExecs = append(orderExecs, currSellOrderExec)
			currSellOrderExec = new(OrderExecution)
		}

		// If the order pubkey is different then we're done with this settlement execution
		if currBuySetExec.Pubkey != currBuy.Order.Pubkey {
			settlementExecs = append(settlementExecs, currBuySetExec)
			currBuySetExec = new(SettlementExecution)
		}
		if currSellSetExec.Pubkey != currSell.Order.Pubkey {
			settlementExecs = append(settlementExecs, currSellSetExec)
			currSellSetExec = new(SettlementExecution)
		}

		// If sell was first, use that price
		if currBuy.Timestamp.UnixNano() > currSell.Timestamp.UnixNano() {
			// otherwise use the buy price
		} else {

		}

	}
	logging.Fatalf("UNIMPLEMENTED!!!")
	return
}

// PrioritizeOrderbookPTP prioritizes orders in a map representation of an orderbook by price-time priority.
// It then separates that into buy and sell lists, which get returned.
// This makes it easy to put in to the MatchPrioritizedOrders algorithm.
func PrioritizeOrderbookPTP(book map[float64][]*LimitOrderIDPair) (buyOrders []*LimitOrderIDPair, sellOrders []*LimitOrderIDPair, err error) {
	logging.Fatalf("UNIMPLEMENTED!!!")
	return
}
