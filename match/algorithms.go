package match

import (
	"fmt"
	"math"
)

// OrderIDPair is a pair of order ID and order, used for generating executions in the matching algorithm
type OrderIDPair struct {
	OrderID [32]byte
	Order   *AuctionOrder
}

// CalculateClearingPrice calculates the clearing price for orders based on their intersections.
// In the future the error will be used potentially since the divide operation at the end might be real bad.
func CalculateClearingPrice(book map[float64][]*OrderIDPair) (clearingPrice float64, err error) {

	// price that is the lowest sell price so far
	var lowestIntersectingPrice float64
	lowestIntersectingPrice = math.MaxFloat64

	// price that is the highest buy price so far
	var highestIntersectingPrice float64
	highestIntersectingPrice = 0

	// Now go through every price in the orderbook, finding the lowest sell order and highest buy order
	for pr, orderPairList := range book {
		for _, orderPair := range orderPairList {
			// make sure that we keep track of the highest buy order price
			if orderPair.Order.IsBuySide() {
				if pr < lowestIntersectingPrice {
					lowestIntersectingPrice = pr
				}
				// make sure we keep track of the lowest sell order price
			} else if orderPair.Order.IsSellSide() {
				if pr > highestIntersectingPrice {
					highestIntersectingPrice = pr
				}
			}
		}
	}

	// fmt.Printf("highestBuy: %f\n", highestIntersectingPrice)
	// fmt.Printf("lowestSell: %f\n", lowestIntersectingPrice)

	// sellClearAmount and buyClearAmount are uint64's, and technically should be amounts of tokens (Issue #22).
	var sellClearAmount uint64
	var buyClearAmount uint64

	// same with totalBuyWant
	var totalBuyWant uint64
	var totalBuyHave uint64

	// same with totalBuyHave
	var totalSellWant uint64
	var totalSellHave uint64
	// now that we have the prices, we go through the book again to calculate the clearing price
	for pr, orderPairList := range book {
		// if there is an intersecting price, calculate clearing amounts for the price.
		for _, orderPair := range orderPairList {
			// for all intersecting prices in the orderbook, we add the amounts
			if orderPair.Order.IsBuySide() && pr >= highestIntersectingPrice {
				buyClearAmount += orderPair.Order.AmountHave
				totalBuyHave += orderPair.Order.AmountHave
				totalBuyWant += orderPair.Order.AmountWant
			} else if orderPair.Order.IsSellSide() && pr <= lowestIntersectingPrice {
				sellClearAmount += orderPair.Order.AmountHave
				totalSellHave += orderPair.Order.AmountHave
				totalSellWant += orderPair.Order.AmountWant
			}
		}
	}

	// fmt.Printf("buyHave: %d\n", buyClearAmount)
	// fmt.Printf("lowestSell: %d\n", sellClearAmount)

	// TODO: this should be changed, I really don't like this floating point math (See Issue #6 and TODOs in match about price.)
	// clearingPrice = (float64(totalBuyWant)/float64(totalBuyHave) + float64(totalSellWant)/float64(totalSellHave)) / 2
	clearingPrice = float64(totalBuyWant+totalSellWant) / float64(totalBuyHave+totalSellHave)
	// fmt.Printf("total buy want: %d\n", totalBuyWant)
	// fmt.Printf("total buy have: %d\n", totalBuyHave)
	// fmt.Printf("total sell want: %d\n", totalSellWant)
	// fmt.Printf("total sell have: %d\n", totalSellHave)
	// fmt.Printf("Clearing price: %f\n", clearingPrice)

	return
}

// ClearingMatchingAlgorithm goes through an orderbook with a clearing price, and generates executions
// based on the clearing matching algorithm
func ClearingMatchingAlgorithm(book map[float64][]*OrderIDPair, clearingPrice float64) (executions []*OrderExecution, err error) {

	var resExec *OrderExecution
	// go through all orders and figure out which ones to match
	for price, orderPairList := range book {
		for _, orderPair := range orderPairList {
			if (orderPair.Order.IsBuySide() && price >= clearingPrice) || (orderPair.Order.IsSellSide() && price <= clearingPrice) {
				// Um so this is needed because of some weird memory issue TODO: remove this fix
				// and put in another fix if you understand pointer black magic
				resExec = new(OrderExecution)
				if *resExec, err = orderPair.Order.GenerateOrderFill(orderPair.OrderID[:], clearingPrice); err != nil {
					err = fmt.Errorf("Error generating execution from clearing price for buy: %s", err)
					return
				}
				executions = append(executions, resExec)
			}
		}
	}

	return
}
