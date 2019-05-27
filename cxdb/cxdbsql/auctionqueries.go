package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// PlaceAuctionPuzzle puts a puzzle and ciphertext in the datastore.
func (db *DB) PlaceAuctionPuzzle(encryptedOrder *match.EncryptedAuctionOrder) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while placing puzzle order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// We don't really care about the result when trying to use a schema
	if _, err = tx.Exec("USE " + db.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Error trying to use auction schema: %s", err)
		return
	}

	var orderBytes []byte
	if orderBytes, err = encryptedOrder.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing order: %s", err)
		return
	}

	// We concatenate ciphertext and puzzle and set "selected" to 1 by default
	placeAuctionPuzzleQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%x', 1);", db.puzzleTable, orderBytes, encryptedOrder.IntendedAuction)
	if _, err = tx.Exec(placeAuctionPuzzleQuery); err != nil {
		err = fmt.Errorf("Error adding auction puzzle to puzzle orders: %s", err)
		return
	}

	return
}

// PlaceAuctionOrder places an order in the unencrypted datastore.
func (db *DB) PlaceAuctionOrder(order *match.AuctionOrder) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while creating new auction: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.auctionSchema + ";"); err != nil {
		err = fmt.Errorf("Error while placing solved auction order: %s", err)
		return
	}

	logging.Infof("Placing order %s!", order)

	// TODO
	return
}

// ViewAuctionOrderBook takes in a trading pair and auction ID, and returns auction orders.
func (db *DB) ViewAuctionOrderBook(tradingPair *match.Pair, auctionID [32]byte) (sellOrderBook []*match.AuctionOrder, buyOrderBook []*match.AuctionOrder, err error) {

	// TODO
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted.
func (db *DB) ViewAuctionPuzzleBook(auctionID [32]byte) (orders []*match.EncryptedAuctionOrder, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while placing puzzle order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// We don't really care about the result when trying to use a schema
	if _, err = tx.Exec("USE " + db.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Error trying to use auction schema: %s", err)
		return
	}

	var rows *sql.Rows
	selectPuzzleQuery := fmt.Sprintf("SELECT encodedOrder FROM %s WHERE auctionID = '%x';", db.puzzleTable, auctionID)
	if rows, err = tx.Query(selectPuzzleQuery); err != nil {
		err = fmt.Errorf("Could not query for puzzles in viewauctionpuzzlebook: %s", err)
		return
	}

	var currEncryptedOrder *match.EncryptedAuctionOrder
	var encodedOrder []byte
	for rows.Next() {
		// allocate memory for the next order to be inserted in the list
		currEncryptedOrder = new(match.EncryptedAuctionOrder)

		if err = rows.Scan(&encodedOrder); err != nil {
			err = fmt.Errorf("Error scanning for puzzle: %s", err)
			return
		}

		// These are all encoded as hex in the db, so decode them
		if _, err = hex.Decode(encodedOrder, encodedOrder); err != nil {
			err = fmt.Errorf("Error decoding puzzle hex returned by database for viewing puzzle book: %s", err)
			return
		}

		if err = currEncryptedOrder.Deserialize(encodedOrder); err != nil {
			err = fmt.Errorf("Error deserializing order stored in db for viewing puzzle book: %s", err)
			return
		}

		// add the order to the list
		orders = append(orders, currEncryptedOrder)

	}

	return
}

// NewAuction takes in an auction ID, and creates a new auction, returning the "height"
// of the auction.
func (db *DB) NewAuction(auctionID [32]byte) (height uint64, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while creating new auction: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// We don't really care about the result when trying to use a schema
	if _, err = tx.Exec("USE " + db.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error trying to use auction order schema: %s", err)
		return
	}

	var nullCompatHeight sql.NullInt64
	auctionNumQuery := fmt.Sprintf("SELECT MAX(auctionNumber) FROM %s;", db.auctionOrderTable)
	if err = tx.QueryRow(auctionNumQuery).Scan(&nullCompatHeight); err != nil {
		err = fmt.Errorf("Could not find maximum auction number when creating new auction: %s", err)
		return
	}

	if nullCompatHeight.Valid {
		height = uint64(nullCompatHeight.Int64)
	} else {
		logging.Warnf("Warning, the max height was null!")
	}

	// Insert the new auction ID w/ current max height + 1
	height++
	insertNewAuctionQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d);", db.auctionOrderTable, auctionID, height)
	if _, err = tx.Exec(insertNewAuctionQuery); err != nil {
		err = fmt.Errorf("Error inserting new auction ID when creating new auction: %s", err)
		return
	}

	return
}

/*
 MatchAuction matches the auction with a specific auctionID. This is meant to be the implementation of pro-rata for just the stuff in the auction. We assume that there are orders in the auction orderbook that are ALL valid.

To understand Pro-rata matching on a batch of orders, here is an example of an orderbook, where the "Buy" list represents all of the buy orders
and the "Sell" list represents all of the sell orders.

The pair is BTC/LTC, you "buy" BTC with LTC and "sell" BTC for LTC.

	"Sell": [
		so4: {
			amountWant: 70 LTC,
			amountHave: 10 BTC,
			// price  : 0.15
		},
		so3: {
			amountWant: 600 LTC,
			amountHave: 100 BTC,
			// price  : 0.17
		},
		so2: {
			amountWant: 400 LTC,
			amountHave: 100 BTC,
			// price  : 0.25
		},
		so1: {
			amountWant: 300 LTC,
			amountHave: 100 BTC,
			// price  : 0.33
		},
	]
	"Buy":  [
		bo4: {
			amountWant: 10 BTC,
			amountHave: 50 LTC,
			// price  : 0.20
		},
		bo3: {
			amountWant: 100 BTC,
			amountHave: 500 LTC,
			// price  : 0.20
		},
		bo2: {
			amountWant: 100 BTC,
			amountHave: 300 LTC,
			// price  : 0.33
		},
		bo1: {
			amountWant: 100 BTC,
			amountHave: 100 LTC,
			// price  : 1.00
		},
	]

We can see here that there's no "nice" way to match these orders, the high/low prices on either end are competitive, nor are there many orders
that are the same price. Pro-rata matching for a single price is trivial.

Alright so we're essentially looking for the orders that will match, meaning we want to create the ordering based on want/have, depending on the price

We start iteratively from the beginning of "Buy" and end of "Sell".

so1 provides 100 BTC, and bo4 is the current order on the buy side.
100 * (portion of 0.20 orders that bo4 provides = 0.1666666666...) > bo4's amountWant, so we match the whole thing, reducing so1 by 10 BTC and 30 LTC. 20 LTC is now left in the LTC pot, since we're not done matching 0.2 orders.

*/
func (db *DB) MatchAuction(auctionID [32]byte) (height uint64, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while creating new auction: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Do this for every pair we have

	// We define the rows, etc here so we don't waste a bunch of stack space
	// ughhhh scanning is so tedious
	var rows *sql.Rows

	var pubkeyBytes []byte
	var auctionIDBytes []byte
	var nonceBytes []byte

	var thisOrder *match.AuctionOrder

	var book map[float64][]*match.AuctionOrder
	book = make(map[float64][]*match.AuctionOrder)

	for _, pair := range db.pairsArray {

		// So we get the orders, and they are supposed to be all valid.
		queryAuctionOrders := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce FROM %s WHERE auctionID='%x';", pair, auctionID)
		if rows, err = tx.Query(queryAuctionOrders); err != nil {
			err = fmt.Errorf("Error querying for orders in auction: %s", err)
			return
		}

		for rows.Next() {
			thisOrder = new(match.AuctionOrder)
			if err = rows.Scan(&pubkeyBytes, &thisOrder.Side, &thisOrder.OrderbookPrice, &thisOrder.AmountHave, &thisOrder.AmountWant, &auctionIDBytes, &nonceBytes); err != nil {
				err = fmt.Errorf("Error scanning row to bytes: %s", err)
				return
			}

			// Now we do the dumb decoding thing
			for _, byteArray := range [][]byte{pubkeyBytes, auctionIDBytes, nonceBytes} {
				// just for everything in this list of things we want to decode from weird words to real bytes
				if _, err = hex.Decode(byteArray, byteArray); err != nil {
					err = fmt.Errorf("Error decoding bytes for auction matching: %s", err)
					return
				}
			}

			// Now we put the data into the auction order
			copy(thisOrder.Pubkey[:], pubkeyBytes)
			copy(thisOrder.AuctionID[:], auctionIDBytes)
			copy(thisOrder.Nonce[:], nonceBytes)

			// okay cool now add it to the list
			book[thisOrder.OrderbookPrice] = append(book[thisOrder.OrderbookPrice], thisOrder)
		}

	}

	return
}
