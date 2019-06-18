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

// MatchPrioritizedOrders matches separated buy and sell orders that are properly sorted in price-time priority
func MatchPrioritizedOrders(buyOrders []*LimitOrderIDPair, sellOrders []*LimitOrderIDPair) (orderExecs []*OrderExecution, settlementExecs []*SettlementExecution, err error) {
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
