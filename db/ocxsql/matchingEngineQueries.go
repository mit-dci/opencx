package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// RunMatchingBestPrices runs matching only on the best prices
func (db *DB) RunMatchingBestPrices(pair match.Pair) (err error) {

	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while updating deposits: \n%s", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while running matching, this might be bad: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	// ======= get sell prices =======

	// First get all the sell prices so we have something to iterate through and match
	getSellPricesQuery := fmt.Sprintf("SELECT DISTINCT price FROM %s WHERE side='%s' ORDER BY price ASC;", pair.String(), "sell")
	sellPriceRows, err := tx.Query(getSellPricesQuery)
	if err != nil {
		return
	}

	var sellPrices []float64

	for sellPriceRows.Next() {
		var sellPrice float64
		if err = sellPriceRows.Scan(&sellPrice); err != nil {
			return
		}

		sellPrices = append(sellPrices, sellPrice)
	}
	if err = sellPriceRows.Close(); err != nil {
		return
	}

	// ======= done with sell prices ======
	// ======= get buy prices =======

	// First get all the buy prices so we have something to iterate through and match
	getBuyPricesQuery := fmt.Sprintf("SELECT DISTINCT price FROM %s WHERE side='%s' ORDER BY price DESC;", pair.String(), "buy")
	buyPriceRows, err := tx.Query(getBuyPricesQuery)
	if err != nil {
		return
	}

	var buyPrices []float64

	for buyPriceRows.Next() {
		var buyPrice float64
		if err = buyPriceRows.Scan(&buyPrice); err != nil {
			return
		}

		buyPrices = append(buyPrices, buyPrice)
	}
	if err = buyPriceRows.Close(); err != nil {
		return
	}

	// ======= done with buy prices ======

	for shouldMatch(buyPrices, sellPrices) {

	}
	return
}

func shouldMatch(buyPrices []float64, sellPrices []float64) bool {
	return len(buyPrices) > 0 && len(sellPrices) > 0 && (buyPrices[0] >= sellPrices[0])
}

// RunMatchingForPriceWithinTransaction runs matching only for a particular price, and takes a transaction
func (db *DB) RunMatchingForPriceWithinTransaction(pair match.Pair, price float64, tx *sql.Tx) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error while running matching for price within transaction, this might be bad: \n%s", err)
			return
		}
	}()

	logging.Infof("Matching all orders with price %f\n", price)

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	// this will select all sell side, ordered by time ascending so the earliest one will be at the front
	getSellSideQuery := fmt.Sprintf("SELECT name, orderID, amountHave, amountWant FROM %s WHERE price=%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "sell")
	sellRows, sellQueryErr := tx.Query(getSellSideQuery)
	if err = sellQueryErr; err != nil {
		return
	}

	var sellOrders []*match.LimitOrder
	for sellRows.Next() {
		sellOrder := new(match.LimitOrder)
		if err = sellRows.Scan(&sellOrder.Client, &sellOrder.OrderID, &sellOrder.AmountHave, &sellOrder.AmountWant); err != nil {
			return
		}

		sellOrders = append(sellOrders, sellOrder)
	}
	if err = sellRows.Close(); err != nil {
		return
	}

	getBuySideQuery := fmt.Sprintf("SELECT name, orderID, amountHave, amountWant FROM %s WHERE price=%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "buy")
	buyRows, buyQueryErr := tx.Query(getBuySideQuery)
	if err = buyQueryErr; err != nil {
		return
	}

	var buyOrders []*match.LimitOrder
	for buyRows.Next() {
		buyOrder := new(match.LimitOrder)
		if err = buyRows.Scan(&buyOrder.Client, &buyOrder.OrderID, &buyOrder.AmountHave, &buyOrder.AmountWant); err != nil {
			return
		}

		buyOrders = append(buyOrders, buyOrder)
	}
	if err = buyRows.Close(); err != nil {
		return
	}

	// loop through them both and make sure there are elements in both otherwise we're good
	for len(buyOrders) > 0 && len(sellOrders) > 0 {
		currBuyOrder := buyOrders[0]
		currSellOrder := sellOrders[0]

		// buying:
		// when we calculate price, could this conditional lead to some weird matching favoritism?
		if currBuyOrder.AmountHave > currSellOrder.AmountWant {

			// keep these to see if we can get any pennies off the order or something?? Isn't that illegal?
			// to see if there's a difference in price technically as well
			prevAmountHave := currSellOrder.AmountHave
			prevAmountWant := currSellOrder.AmountWant
			currBuyOrder.AmountHave -= currSellOrder.AmountWant
			currBuyOrder.AmountWant -= currSellOrder.AmountHave

			// update order with new amounts
			if err = db.UpdateOrderAmountsWithinTransaction(currBuyOrder, pair, tx); err != nil {
				return
			}
			// delete sell order
			if err = db.DeleteOrderWithinTransaction(currSellOrder, pair, tx); err != nil {
				return
			}

			sellOrders = sellOrders[1:]

			// use the balance schema because we're ending with balance transactions
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Client, prevAmountHave, tx, pair.AssetWant.GetAssociatedCoinParam()); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Client, prevAmountWant, tx, pair.AssetHave.GetAssociatedCoinParam()); err != nil {
				return
			}

			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		} else if currBuyOrder.AmountHave < currSellOrder.AmountWant {

			// keep these to see if we can get any pennies off the order or something?? Isn't that illegal?
			// to see if there's a difference in price technically as well
			prevAmountHave := currBuyOrder.AmountHave
			prevAmountWant := currBuyOrder.AmountWant
			currSellOrder.AmountHave -= currBuyOrder.AmountWant
			currSellOrder.AmountWant -= currBuyOrder.AmountHave

			// update order with new amounts
			if err = db.UpdateOrderAmountsWithinTransaction(currSellOrder, pair, tx); err != nil {
				return
			}
			// delete buy order
			if err = db.DeleteOrderWithinTransaction(currBuyOrder, pair, tx); err != nil {
				return
			}

			buyOrders = buyOrders[1:]
			// use the balance schema because we're ending with balance transactions
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Client, prevAmountWant, tx, pair.AssetWant.GetAssociatedCoinParam()); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Client, prevAmountHave, tx, pair.AssetHave.GetAssociatedCoinParam()); err != nil {
				return
			}

			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		} else if currBuyOrder.AmountHave == currSellOrder.AmountWant {

			// this is if they can perfectly fill each others orders

			// delete buy order
			if err = db.DeleteOrderWithinTransaction(currBuyOrder, pair, tx); err != nil {
				return
			}
			// delete sell order
			if err = db.DeleteOrderWithinTransaction(currSellOrder, pair, tx); err != nil {
				return
			}

			sellOrders = sellOrders[1:]
			buyOrders = buyOrders[1:]

			// use the balance schema because we're ending with balance transactions
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Client, currBuyOrder.AmountWant, tx, pair.AssetWant.GetAssociatedCoinParam()); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Client, currBuyOrder.AmountHave, tx, pair.AssetHave.GetAssociatedCoinParam()); err != nil {
				return
			}

			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		}
	}

	return
}

// RunMatching is public because it's cool to run it at any time
func (db *DB) RunMatching(pair match.Pair) (err error) {

	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while updating deposits: \n%s", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while running matching, this might be bad: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	// First get all the prices so we have something to iterate through and match
	getPricesQuery := fmt.Sprintf("SELECT DISTINCT price FROM %s;", pair.String())
	rows, err := tx.Query(getPricesQuery)
	if err != nil {
		return
	}

	var prices []float64

	for rows.Next() {
		var price float64
		if err = rows.Scan(&price); err != nil {
			return
		}

		prices = append(prices, price)
	}
	if err = rows.Close(); err != nil {
		return
	}

	for _, price := range prices {

		if err = db.RunMatchingForPriceWithinTransaction(pair, price, tx); err != nil {
			return
		}

	}

	return
}
