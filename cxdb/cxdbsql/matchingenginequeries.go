package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/opencx/util"

	"github.com/mit-dci/lit/coinparam"

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

	var currSide string
	var currPrice float64
	var currAmountHave uint64
	var currAmountWant uint64
	for rows.Next() {
		currSide = *new(string)
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

// RunMatchingBestPrices runs matching only on the best prices. Creates a transaction.
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

	if err = db.RunMatchingBestPricesWithinTransaction(pair, tx); err != nil {
		return
	}

	return
}

// RunMatchingBestPricesWithinTransaction matches the best prices within a transaction.
func (db *DB) RunMatchingBestPricesWithinTransaction(pair *match.Pair, tx *sql.Tx) (err error) {
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

	// this is a really really basic / naive algorithm that should run matching for the "best price"
	if shouldMatch(buyPrices, sellPrices) {
		if buyPrices[0] == sellPrices[0] {
			// If they're the same then great we can just match all this stuff at price
			if err = db.RunMatchingForPriceWithinTransaction(pair, buyPrices[0], tx); err != nil {
				return
			}
		} else {
			// otherwise we have to figure out the crossed trades
			// this situation puts us at a point where the lowest sell price is less than the highest buy price
			// if the lower sell order was placed before the high buy order then the trade gets executed at the sell order price
			// if the higher buy order was placed before the low sell order then the trade gets executed at the buy order price

			if err = db.RunMatchingCrossedPricesWithinTransaction(pair, buyPrices[0], sellPrices[0], tx); err != nil {
				return
			}
		}
	}

	return
}

func shouldMatch(buyPrices []float64, sellPrices []float64) bool {
	return len(buyPrices) > 0 && len(sellPrices) > 0 && (buyPrices[0] >= sellPrices[0])
}

// RunMatchingCrossedPricesWithinTransaction runs the matching algorithm assuming the bestBuyPrice is actually the best buy price and the bestSellPrice is actually the best sell price
func (db *DB) RunMatchingCrossedPricesWithinTransaction(pair *match.Pair, bestBuyPrice float64, bestSellPrice float64, tx *sql.Tx) (err error) {

	// get coinparam for assetwant
	var assetWantCoinType *coinparam.Params
	if assetWantCoinType, err = util.GetParamFromName(pair.AssetWant.String()); err != nil {
		err = fmt.Errorf("Tried to run matching for asset that doesn't have a coinType. Nothing will be compatible")
		return
	}

	// get coinparam for assetwant
	var assetHaveCoinType *coinparam.Params
	if assetHaveCoinType, err = util.GetParamFromName(pair.AssetHave.String()); err != nil {
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
	getSellSideQuery := fmt.Sprintf("SELECT pubkey, orderID, amountHave, amountWant FROM %s WHERE price>=%f AND side='%s' ORDER BY time ASC;", pair.String(), bestSellPrice, "sell")
	sellRows, sellQueryErr := tx.Query(getSellSideQuery)
	if err = sellQueryErr; err != nil {
		return
	}

	var sellOrders []*match.LimitOrder
	for sellRows.Next() {
		var pubkeyBytes []byte
		sellOrder := new(match.LimitOrder)
		if err = sellRows.Scan(&pubkeyBytes, &sellOrder.OrderID, &sellOrder.AmountHave, &sellOrder.AmountWant); err != nil {
			return
		}

		// we have to do this because ugh they return my byte arrays as hex strings...
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			return
		}

		copy(sellOrder.Pubkey[:], pubkeyBytes)
		sellOrders = append(sellOrders, sellOrder)
	}
	if err = sellRows.Close(); err != nil {
		return
	}

	// logging.Infof("Sell orders length: %d", len(sellOrders))
	getBuySideQuery := fmt.Sprintf("SELECT pubkey, orderID, amountHave, amountWant FROM %s WHERE price<=%f AND side='%s' ORDER BY time ASC;", pair.String(), bestBuyPrice, "buy")
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

		// we have to do this because ugh they return my byte arrays as hex strings...
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			return
		}

		copy(buyOrder.Pubkey[:], pubkeyBytes)
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

			var buyOrderPubkey *koblitz.PublicKey
			if buyOrderPubkey, err = koblitz.ParsePubKey(currBuyOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			var sellOrderPubkey *koblitz.PublicKey
			if sellOrderPubkey, err = koblitz.ParsePubKey(currSellOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.AddToBalanceWithinTransaction(buyOrderPubkey, prevAmountHave, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.AddToBalanceWithinTransaction(sellOrderPubkey, prevAmountWant, tx, assetHaveCoinType); err != nil {
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

			var buyOrderPubkey *koblitz.PublicKey
			if buyOrderPubkey, err = koblitz.ParsePubKey(currBuyOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			var sellOrderPubkey *koblitz.PublicKey
			if sellOrderPubkey, err = koblitz.ParsePubKey(currSellOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.AddToBalanceWithinTransaction(buyOrderPubkey, prevAmountWant, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.AddToBalanceWithinTransaction(sellOrderPubkey, prevAmountHave, tx, assetHaveCoinType); err != nil {
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

			var buyOrderPubkey *koblitz.PublicKey
			if buyOrderPubkey, err = koblitz.ParsePubKey(currBuyOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			var sellOrderPubkey *koblitz.PublicKey
			if sellOrderPubkey, err = koblitz.ParsePubKey(currSellOrder.Pubkey[:], koblitz.S256()); err != nil {
				return
			}

			// credit buyOrder client with sellOrder amountHave
			if err = db.AddToBalanceWithinTransaction(buyOrderPubkey, currBuyOrder.AmountWant, tx, assetWantCoinType); err != nil {
				return
			}
			// credit sellOrder client with buyorder amountWant
			if err = db.AddToBalanceWithinTransaction(sellOrderPubkey, currBuyOrder.AmountHave, tx, assetHaveCoinType); err != nil {
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

// RunMatchingForPriceWithinTransaction runs matching only for a particular price, and takes a transaction
func (db *DB) RunMatchingForPriceWithinTransaction(pair *match.Pair, price float64, tx *sql.Tx) (err error) {

	if err = db.RunMatchingCrossedPricesWithinTransaction(pair, price, price, tx); err != nil {
		return
	}

	return
}
