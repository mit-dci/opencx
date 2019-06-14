package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/mit-dci/opencx/util"
	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

// PlaceOrder runs the queries which places an input order. Placing an individual order is atomic.
func (db *DB) PlaceOrder(order *match.LimitOrder) (orderid string, err error) {

	// Check that they have the balance for the order
	// if they do, place the order and update their balance

	// hash order so we can use that as a primary key
	sha := sha3.New256()
	var orderBytes []byte
	if orderBytes, err = order.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing while placing order: %s", err)
		return
	}
	sha.Write(orderBytes)
	hashedOrder := sha.Sum(nil)

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
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

	inPairs := db.doesPairExist(&order.TradingPair)
	if !inPairs {
		err = fmt.Errorf("Trading pair does not exist, try the other way around (e.g. ltc/btc => btc/ltc)")
		return
	}

	var balCheckAsset string
	if order.Side == match.Buy {
		balCheckAsset = order.TradingPair.AssetHave.String()
	} else {
		balCheckAsset = order.TradingPair.AssetWant.String()
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", balCheckAsset, order.Pubkey[:])
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
			subtractBalanceQuery := fmt.Sprintf("UPDATE %s SET balance=%d WHERE pubkey='%x';", balCheckAsset, newBal, order.Pubkey[:])
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

			placeOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s', '%s', %f, %d, %d, NOW());", order.TradingPair.String(), order.Pubkey[:], order.Side, realPrice, order.AmountHave, order.AmountWant)
			if _, err = tx.Exec(placeOrderQuery); err != nil {
				return
			}

		} else {
			err = fmt.Errorf("Tried to place an order for more than you own, please lower the amount you want or adjust price")
			return
		}
	} else {
		err = fmt.Errorf("Could not find balance for user with pubkey %x, so cannot place order", order.Pubkey[:])
		return
	}

	orderid = order.OrderID

	if err = db.RunMatchingBestPricesWithinTransaction(&order.TradingPair, tx); err != nil {
		return
	}

	return
}

// UpdateOrderAmountsWithinTransaction updates a single order within a sql transaction. It takes the order ID from the order passed
// in and tries to update the amounts as specified in the order passed.
func (db *DB) UpdateOrderAmountsWithinTransaction(order *match.LimitOrder, pair *match.Pair, tx *sql.Tx) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating order within transaction: \n%s", err)
			return
		}
	}()

	updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE orderID='%x';", pair.String(), order.AmountHave, order.AmountWant, order.OrderID)
	if _, err = tx.Exec(updateOrderQuery); err != nil {
		err = fmt.Errorf("Error updating order within transaction: %s", err)
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

	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%x';", pair.String(), order.OrderID)
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

	didOrderExist := false
	for _, pair := range db.pairsArray {
		// figure out if there is even an order
		getCurrentOrderQuery := fmt.Sprintf("SELECT pubkey, amountHave, amountWant, side FROM %s WHERE orderID='%x';", pair.String(), orderID)
		rows, currOrderErr := tx.Query(getCurrentOrderQuery)
		if err = currOrderErr; err != nil {
			return
		}

		if rows.Next() {
			var pubkeyBytes []byte
			var amtHave uint64
			var amtWant uint64
			var side string

			didOrderExist = true

			// get current values in case of partially filled order

			if err = rows.Scan(&pubkeyBytes, &amtHave, &amtWant, &side); err != nil {
				return
			}

			var pubkey *koblitz.PublicKey
			if pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
				return
			}

			// do this so we don't get bad connection / busy buffer issues
			if err = rows.Close(); err != nil {
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
				// make sure that the asset has an associated coinparam. Don't trade assets that you can't settle on a ledger.
				if correctAssetHave, err = util.GetParamFromName(pair.AssetHave.String()); err != nil {
					return
				}
			} else if side == "sell" {
				// make sure that the asset has an associated coinparam. Don't trade assets that you can't settle on a ledger.
				if correctAssetHave, err = util.GetParamFromName(pair.AssetWant.String()); err != nil {
					return
				}
			}

			// update the balance of the client
			if err = db.AddToBalanceWithinTransaction(pubkey, amtHave, tx, correctAssetHave); err != nil {
				return
			}
		}
		// use order schema at end of loop so we go back to where we were
		if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
			return
		}
	}
	if !didOrderExist {
		err = fmt.Errorf("Order does not exist in any orderbook")
		return
	}

	// credit client with amounthave
	return
}

// ViewLimitOrderBook takes in a trading pair, and returns limit orders.
func (db *DB) ViewLimitOrderBook(tradingPair *match.Pair) (orderbook map[float64][]*match.LimitOrderIDPair, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for ViewLimitOrderBook: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while viewing auction order book: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	if orderbook, err = db.ViewLimitOrderBookTx(tradingPair, tx); err != nil {
		err = fmt.Errorf("Error viewing auction orderbook for tx: %s", err)
		return
	}

	return
}

func (db *DB) ViewLimitOrderBookTx(tradingPair *match.Pair, tx *sql.Tx) (orderbook map[float64][]*match.LimitOrderIDPair, err error) {

	orderbook = make(map[float64][]*match.LimitOrderIDPair)
	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for viewlimitorderbook: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, orderID FROM %s;", tradingPair)
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting orders from db for viewlimitorderbook: %s", err)
		return
	}

	defer func() {
		// TODO: if there's a better way to chain all these errors, figure it out
		var newErr error
		if newErr = rows.Close(); newErr != nil {
			err = fmt.Errorf("Error closing rows for viewlimitorderbook: %s", newErr)
			return
		}
		return
	}()

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.LimitOrder
	var thisOrderPair *match.LimitOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64

	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.LimitOrder)
		thisOrderPair = new(match.LimitOrderIDPair)
		if err = rows.Scan(&pkBytes, &thisOrder.Side, &thisPrice, &thisOrder.AmountHave, &thisOrder.AmountWant, &hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error scanning into order for viewlimitorderbook: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		for _, byteArrayPtr := range []*[]byte{&pkBytes, &hashedOrderBytes} {
			if *byteArrayPtr, err = hex.DecodeString(string(*byteArrayPtr)); err != nil {
				err = fmt.Errorf("Error decoding bytes for viewlimitorderbook: %s", err)
				return
			}
		}

		// Copy all of the bytes
		copy(thisOrder.Pubkey[:], pkBytes)
		copy(thisOrderPair.OrderID[:], hashedOrderBytes)

		orderbook[thisPrice] = append(orderbook[thisPrice], thisOrderPair)
		thisOrderPair.Order = thisOrder

	}

	return
}

// GetOrder gets an order for a specific order ID
func (db *DB) GetOrder(orderID string) (order *match.LimitOrder, err error) {
	tx, err := db.DBHandler.Begin()
	if err != nil {
		err = fmt.Errorf("Error getting order: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.orderSchema + ";"); err != nil {
		return
	}

	for _, pair := range db.pairsArray {
		// figure out if there is even an order
		getCurrentOrderQuery := fmt.Sprintf("SELECT pubkey, amountHave, amountWant, side, timestamp FROM %s WHERE orderID='%s';", pair.String(), orderID)
		rows, currOrderErr := tx.Query(getCurrentOrderQuery)
		if err = currOrderErr; err != nil {
			return
		}

		if rows.Next() {
			var pubkeyBytes []byte
			if err = rows.Scan(&pubkeyBytes, &order.AmountHave, &order.AmountWant, &order.Side, &order.Timestamp); err != nil {
				return
			}

			// we have to do this because ugh they return my byte arrays as hex strings...
			if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
				return
			}

			// copy pubkey bytes into 33 order byte arr
			copy(order.Pubkey[:], pubkeyBytes)

			order.TradingPair = *pair
			order.OrderID = orderID

		}
		// do this so we don't get bad connection / busy buffer issues
		if err = rows.Close(); err != nil {
			return
		}
	}

	if order != nil {
		err = fmt.Errorf("Order does not exist in any orderbook")
		return
	}

	return
}

// GetOrdersForPubkey gets orders for a specific pubkey
func (db *DB) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.LimitOrderIDPair, err error) {
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

	orders = make(map[float64][]*match.LimitOrderIDPair)
	var currPrice float64

	for _, pair := range db.pairsArray {
		// figure out if there is even an order
		getCurrentOrderQuery := fmt.Sprintf("SELECT orderID, amountHave, amountWant, side, timestamp, price FROM %s WHERE pubkey='%x';", pair.String(), pubkey.SerializeCompressed())
		rows, currOrderErr := tx.Query(getCurrentOrderQuery)
		if err = currOrderErr; err != nil {
			return
		}

		var orderpair *match.LimitOrderIDPair
		var order *match.LimitOrder
		for rows.Next() {
			order = new(match.LimitOrder)
			orderpair = new(match.LimitOrderIDPair)

			if err = rows.Scan(&order.OrderID, &order.AmountHave, &order.AmountWant, &order.Side, &order.Timestamp, &currPrice); err != nil {
				return
			}

			// copy pubkey bytes into 33 order byte arr
			copy(order.Pubkey[:], pubkey.SerializeCompressed())
			order.TradingPair = *pair

			orders[currPrice] = append(orders[currPrice], order)

		}
		// do this so we don't get bad connection / busy buffer issues
		if err = rows.Close(); err != nil {
			return
		}
	}

	return
}
