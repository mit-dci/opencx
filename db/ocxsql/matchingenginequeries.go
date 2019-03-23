package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/opencx/util"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// RunMatchingForPrice runs matching for only a specific price. Creates a transaction.
func (db *DB) RunMatchingForPrice(pair *match.Pair, price float64) (err error) {

	// create the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		err = fmt.Errorf("Error beginning transaction while running matching for price: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error running matching for price, this might be bad: \n%s", err)
		}
		err = tx.Commit()
	}()

	if err = db.RunMatchingForPriceWithinTransaction(pair, price, tx); err != nil {
		logging.Errorf("This is the error runmatchpricewthintx is getting: %s", err)
		return
	}

	return
}

// CalculatePrice calculates the price based on the volume and side of the orders.
func (db *DB) CalculatePrice(pair *match.Pair) (price float64, err error) {

	// create the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		err = fmt.Errorf("Error beginning transaction while running matching for price: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error calculating price, this might be bad: \n%s", err)
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	var sellExpectation float64
	var buyExpectation float64
	var totalVolume uint64

	getAllOrdersQuery := fmt.Sprintf("SELECT side, price, amountHave, amountWant FROM %s;", pair.String())
	rows, err := tx.Query(getAllOrdersQuery)
	if err != nil {
		return
	}

	for rows.Next() {
		var currSide string
		var currPrice float64
		var currAmountHave uint64
		var currAmountWant uint64
		if err = rows.Scan(&currSide, &currPrice, &currAmountHave, &currAmountWant); err != nil {
			return
		}

		if currSide == "buy" {
			buyExpectation += float64(currAmountHave) * currPrice
			totalVolume += currAmountHave
		} else if currSide == "sell" {
			sellExpectation += float64(currAmountWant) * currPrice
			totalVolume += currAmountWant
		}

	}

	if err = rows.Close(); err != nil {
		return
	}

	price = (sellExpectation + buyExpectation) / float64(totalVolume)

	return
}

// RunMatchingBestPrices runs matching only on the best prices
func (db *DB) RunMatchingBestPrices(pair *match.Pair) (err error) {

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

	var lastBuyPrice float64
	var lastSellPrice float64

	// this is a really really basic / naive algorithm that should run matching for the "best price"
	for shouldMatch(buyPrices, sellPrices) {
		if buyPrices[0] == sellPrices[0] {
			if err = db.RunMatchingForPriceWithinTransaction(pair, buyPrices[0], tx); err != nil {
				return
			}
		} else {
			if err = db.RunMatchingForPriceWithinTransaction(pair, buyPrices[0], tx); err != nil {
				return
			}
			if err = db.RunMatchingForPriceWithinTransaction(pair, sellPrices[0], tx); err != nil {
				return
			}
		}
		lastBuyPrice = buyPrices[0]
		lastSellPrice = sellPrices[0]
		buyPrices = buyPrices[1:]
		sellPrices = sellPrices[1:]
	}

	if lastBuyPrice > 0 && lastSellPrice > 0 {
		midpoint := (lastBuyPrice + lastSellPrice) / 2
		db.SetPrice(midpoint, pair.String())
		logging.Infof("Set price to %f for %s\n", midpoint, pair.String())
	}
	return
}

func shouldMatch(buyPrices []float64, sellPrices []float64) bool {
	return len(buyPrices) > 0 && len(sellPrices) > 0 && (buyPrices[0] >= sellPrices[0])
}

// RunMatchingForPriceWithinTransaction runs matching only for a particular price, and takes a transaction
func (db *DB) RunMatchingForPriceWithinTransaction(pair *match.Pair, price float64, tx *sql.Tx) (err error) {
	// get coinparam for assetwant
	var assetWantCoinType *coinparam.Params
	if assetWantCoinType, err = util.GetCoinTypeFromName(pair.AssetWant.String()); err != nil {
		err = fmt.Errorf("Tried to run matching for asset that doesn't have a coinType. Nothing will be compatible")
		return
	}

	// get coinparam for assetwant
	var assetHaveCoinType *coinparam.Params
	if assetHaveCoinType, err = util.GetCoinTypeFromName(pair.AssetHave.String()); err != nil {
		err = fmt.Errorf("Tried to run matching for asset that doesn't have a coinType. Nothing will be compatible")
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error while running matching for price within transaction, this might be bad: \n%s", err)
			return
		}
	}()

	// debug
	// logging.Infof("Matching all orders with price %f\n", price)

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	// this will select all sell side, ordered by time ascending so the earliest one will be at the front
	getSellSideQuery := fmt.Sprintf("SELECT pubkey, orderID, amountHave, amountWant FROM %s WHERE price>%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "sell")
	sellRows, sellQueryErr := tx.Query(getSellSideQuery)
	if err = sellQueryErr; err != nil {
		return
	}

	var sellOrders []*match.LimitOrder
	for sellRows.Next() {
		sellOrder := new(match.LimitOrder)
		var pubkeyBytes []byte
		if err = sellRows.Scan(&pubkeyBytes, &sellOrder.OrderID, &sellOrder.AmountHave, &sellOrder.AmountWant); err != nil {
			return
		}

		if sellOrder.Pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
			return
		}

		sellOrders = append(sellOrders, sellOrder)
	}
	if err = sellRows.Close(); err != nil {
		return
	}

	// logging.Infof("Sell orders length: %d", len(sellOrders))
	getBuySideQuery := fmt.Sprintf("SELECT pubkey, orderID, amountHave, amountWant FROM %s WHERE price<%f AND side='%s' ORDER BY time ASC;", pair.String(), price, "buy")
	buyRows, buyQueryErr := tx.Query(getBuySideQuery)
	if err = buyQueryErr; err != nil {
		return
	}

	var buyOrders []*match.LimitOrder
	for buyRows.Next() {
		var pubkeyBytes []byte
		buyOrder := new(match.LimitOrder)
		if err = buyRows.Scan(&pubkeyBytes, &buyOrder.OrderID, &buyOrder.AmountHave, &buyOrder.AmountWant); err != nil {
			return
		}

		if buyOrder.Pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
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

		if currBuyOrder.AmountHave > currSellOrder.AmountWant {

			prevAmountHave := currSellOrder.AmountHave
			prevAmountWant := currSellOrder.AmountWant

			// this partial fulfillment / uint underflow quick fix needs to be looked into. Are we losing any money here?
			if currBuyOrder.AmountWant < currSellOrder.AmountHave {
				currBuyOrder.AmountWant = 0
				logging.Infof("Underflow encountered. Difference in %d satoshis of %s", currSellOrder.AmountHave-currBuyOrder.AmountWant, pair.AssetWant)
			} else {
				currBuyOrder.AmountWant -= currSellOrder.AmountHave
			}
			currBuyOrder.AmountHave -= currSellOrder.AmountWant

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
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Pubkey, prevAmountHave, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Pubkey, prevAmountWant, tx, assetHaveCoinType); err != nil {
				return
			}

			// logging.Infof("done all greater")
			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		} else if currBuyOrder.AmountHave < currSellOrder.AmountWant {

			prevAmountHave := currBuyOrder.AmountHave
			prevAmountWant := currBuyOrder.AmountWant

			// this partial fulfillment / uint underflow quick fix needs to be looked into. Are we losing any money here?
			if currSellOrder.AmountHave < currBuyOrder.AmountWant {
				currSellOrder.AmountHave = 0
				logging.Infof("Underflow encountered. Difference in %d satoshis of %s", currBuyOrder.AmountWant-currSellOrder.AmountHave, pair.AssetWant)
			} else {
				currSellOrder.AmountHave -= currBuyOrder.AmountWant
			}
			currSellOrder.AmountWant -= currBuyOrder.AmountHave

			// logging.Infof("less")
			// update order with new amounts
			if err = db.UpdateOrderAmountsWithinTransaction(currSellOrder, pair, tx); err != nil {
				return
			}
			// delete buy order
			if err = db.DeleteOrderWithinTransaction(currBuyOrder, pair, tx); err != nil {
				return
			}

			// logging.Infof("done delete")
			buyOrders = buyOrders[1:]
			// use the balance schema because we're ending with balance transactions
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Pubkey, prevAmountWant, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Pubkey, prevAmountHave, tx, assetHaveCoinType); err != nil {
				return
			}

			// logging.Infof("done lesser")
			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		} else if currBuyOrder.AmountHave == currSellOrder.AmountWant {

			// this is if they can perfectly fill each others orders

			// logging.Infof("Order amounts are equal")
			// delete buy order
			if err = db.DeleteOrderWithinTransaction(currBuyOrder, pair, tx); err != nil {
				return
			}
			// delete sell order
			if err = db.DeleteOrderWithinTransaction(currSellOrder, pair, tx); err != nil {
				return
			}

			// logging.Infof("deleted orders")
			sellOrders = sellOrders[1:]
			buyOrders = buyOrders[1:]

			// use the balance schema because we're ending with balance transactions
			if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.UpdateBalanceWithinTransaction(currBuyOrder.Pubkey, currBuyOrder.AmountWant, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.UpdateBalanceWithinTransaction(currSellOrder.Pubkey, currBuyOrder.AmountHave, tx, assetHaveCoinType); err != nil {
				return
			}

			// logging.Infof("done update")

			// making sure we're going back in the order db, we're going to be making lots of order queries
			if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
				return
			}
		}
	}

	return
}

// RunMatching runs matching on every price in the order book. If you had enough processing power, this would be the matching to
// run, since it scans all prices, and can be run at any time.
func (db *DB) RunMatching(pair *match.Pair) (err error) {

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
