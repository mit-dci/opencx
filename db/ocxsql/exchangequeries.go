package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlaceOrder runs the queries which places an input order. Placing an individual order is atomic.
func (db *DB) PlaceOrder(order *match.LimitOrder) (orderid string, err error) {

	// Check that they have the balance for the order
	// if they do, place the order and update their balance
	err = order.SetID()
	if err != nil {
		return
	}

	tx, err := db.DBHandler.Begin()
	if err != nil {
		err = fmt.Errorf("Error beginning transaction while placing order: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while placing order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use balance schema
	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	inPairs := false
	for _, pair := range db.pairsArray {
		inPairs = inPairs || pair.String() == order.TradingPair.String()
	}
	if !inPairs {
		err = fmt.Errorf("Trading pair does not exist, try the other way around (e.g. ltc/btc => btc/ltc)")
		return
	}

	var balCheckAsset string
	if order.IsBuySide() {
		balCheckAsset = order.TradingPair.AssetHave.String()
	} else {
		balCheckAsset = order.TradingPair.AssetWant.String()
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", balCheckAsset, order.Client)
	rows, getBalErr := tx.Query(getBalanceQuery)
	if err = getBalErr; err != nil {
		return
	}

	if rows.Next() {
		var balance uint64
		if err = rows.Scan(&balance); err != nil {
			return
		}

		if err = rows.Close(); err != nil {
			return
		}

		if balance > order.AmountHave {

			newBal := balance - order.AmountHave
			subtractBalanceQuery := fmt.Sprintf("UPDATE %s SET balance=%d WHERE name='%s';", balCheckAsset, newBal, order.Client)
			if _, err = tx.Query(subtractBalanceQuery); err != nil {
				return
			}

			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}

			realPrice, priceErr := order.Price()
			if err = priceErr; err != nil {
				return
			}

			//debug
			// logging.Infof("Price for %s side: %f\n", order.Side, realPrice)

			placeOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', %f, %d, %d, NOW());", order.TradingPair.String(), order.Client, order.OrderID, order.Side, realPrice, order.AmountHave, order.AmountWant)
			if _, err = tx.Exec(placeOrderQuery); err != nil {
				return
			}

		} else {
			err = fmt.Errorf("Tried to place an order for more than you own, please lower the amount you want or adjust price")
			return
		}
	} else {
		err = fmt.Errorf("Could not find balance for user %s, so cannot place order", order.Client)
		return
	}

	//debug
	// logging.Infof("done")
	orderid = order.OrderID

	// when placing an order subtract from the balance
	return
}

// UpdateOrderAmountsWithinTransaction updates a single order within a sql transaction
func (db *DB) UpdateOrderAmountsWithinTransaction(order *match.LimitOrder, pair *match.Pair, tx *sql.Tx) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating order within transaction: \n%s", err)
			return
		}
	}()

	updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE orderID='%s';", pair.String(), order.AmountHave, order.AmountWant, order.OrderID)
	if _, err = tx.Exec(updateOrderQuery); err != nil {
		logging.Infof("weird order thing: %s", updateOrderQuery)
		return
	}
	return
}

// DeleteOrderWithinTransaction deletes an order within a transaction.
func (db *DB) DeleteOrderWithinTransaction(order *match.LimitOrder, pair *match.Pair, tx *sql.Tx) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error deleting order within transaction: \n%s", err)
			return
		}
	}()

	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%s';", pair.String(), order.OrderID)
	if _, err = tx.Exec(deleteOrderQuery); err != nil {
		return
	}
	return
}

// CancelOrder runs the queries to cancel an order. Cancelling an individual order is atomic.
func (db *DB) CancelOrder(orderID string) (err error) {

	tx, err := db.DBHandler.Begin()
	if err != nil {
		err = fmt.Errorf("Error cancelling order: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while cancelling order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	for _, pair := range db.pairsArray {
		// figure out if there is even an order
		getCurrentOrderQuery := fmt.Sprintf("SELECT name, amountHave, amountWant, side FROM %s WHERE orderID='%s';", pair.String(), orderID)
		rows, currOrderErr := tx.Query(getCurrentOrderQuery)
		if err = currOrderErr; err != nil {
			return
		}

		if rows.Next() {
			var client string
			var amtHave uint64
			var amtWant uint64
			var side string

			// get current values in case of partially filled order
			err = rows.Scan(&client, &amtHave, &amtWant, &side)
			if err != nil {
				return
			}

			// delete order from db
			deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%s';", pair.String(), orderID)
			if _, err = tx.Exec(deleteOrderQuery); err != nil {
				return
			}

			// use balance schema for updating balance
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			var correctAssetHave *coinparam.Params
			if side == "buy" {
				correctAssetHave = pair.AssetHave.GetAssociatedCoinParam()
			} else if side == "sell" {
				correctAssetHave = pair.AssetWant.GetAssociatedCoinParam()
			}

			// update the balance of the client
			if err = db.UpdateBalanceWithinTransaction(client, amtHave, tx, correctAssetHave); err != nil {
				return
			}
		}
	}

	// credit client with amounthave
	return
}

// ViewOrderBook returns a list of orders that is the orderbook
func (db *DB) ViewOrderBook(pair *match.Pair) (sellOrderBook []*match.LimitOrder, buyOrderBook []*match.LimitOrder, err error) {

	tx, err := db.DBHandler.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("Error beginning transaction while updating deposits: \n%s", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while viewing order book: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	// First get all the prices so we have something to iterate through and match
	getPricesQuery := fmt.Sprintf("SELECT DISTINCT price FROM %s ORDER BY price ASC;", pair.String())
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

		// logging.Infof("Matching all orders with price %f\n", price)

		if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
			return
		}

		// this will select all sell side, ordered by time ascending so the earliest one will be at the front
		getSellSideQuery := fmt.Sprintf("SELECT name, orderID, side, amountHave, amountWant FROM %s WHERE price=%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "sell")
		sellRows, sellQueryErr := tx.Query(getSellSideQuery)
		if err = sellQueryErr; err != nil {
			return
		}

		var sellOrders []*match.LimitOrder
		for sellRows.Next() {
			sellOrder := new(match.LimitOrder)
			if err = sellRows.Scan(&sellOrder.Client, &sellOrder.OrderID, &sellOrder.Side, &sellOrder.AmountHave, &sellOrder.AmountWant); err != nil {
				return
			}
			// set price to return to clients
			sellOrder.OrderbookPrice = price

			sellOrders = append(sellOrders, sellOrder)
		}
		if err = sellRows.Close(); err != nil {
			return
		}

		sellOrderBook = append(sellOrderBook, sellOrders[:]...)
		getBuySideQuery := fmt.Sprintf("SELECT name, orderID, side, amountHave, amountWant FROM %s WHERE price=%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "buy")
		buyRows, buyQueryErr := tx.Query(getBuySideQuery)
		if err = buyQueryErr; err != nil {
			return
		}

		var buyOrders []*match.LimitOrder
		for buyRows.Next() {
			buyOrder := new(match.LimitOrder)
			if err = buyRows.Scan(&buyOrder.Client, &buyOrder.OrderID, &buyOrder.Side, &buyOrder.AmountHave, &buyOrder.AmountWant); err != nil {
				return
			}
			// set price to return to clients
			buyOrder.OrderbookPrice = price

			buyOrders = append(buyOrders, buyOrder)
		}
		if err = buyRows.Close(); err != nil {
			return
		}

		buyOrderBook = append(buyOrderBook, buyOrders[:]...)

	}

	return
}
