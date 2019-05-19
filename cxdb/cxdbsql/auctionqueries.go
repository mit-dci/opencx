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

// MatchAuction matches the auction with a specific auctionID. This is meant to be the implementation of pro-rata for just the stuff in the auction. We assume that there are orders in the auction orderbook that are ALL valid.
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

	// I hate writing matching algorithms!!!!!
	// Alright so here's what's going to happen:
	// uh two lists? How does crossing work with pro rata
	// very annoying, maybe it's time to look at other examples

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
