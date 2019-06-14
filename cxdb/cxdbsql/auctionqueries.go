package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
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

// PlaceAuctionOrder places an order in the unencrypted datastore. This assumes that the order is valid.
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
		return
	}()

	if _, err = tx.Exec("USE " + db.auctionSchema + ";"); err != nil {
		err = fmt.Errorf("Error while placing solved auction order: %s", err)
		return
	}

	logging.Infof("Placing order %s!", order)

	// calculate price
	var price float64
	if price, err = order.Price(); err != nil {
		err = fmt.Errorf("Error getting price from order while placing order: %s", err)
		return
	}

	// hash order so we can use that as a primary key
	sha := sha3.New256()
	sha.Write(order.SerializeSignable())
	hashedOrder := sha.Sum(nil)

	insertOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s', %f, %d, %d, '%x', '%x', '%x', '%x');", &order.TradingPair, order.Pubkey, order.Side, price, order.AmountHave, order.AmountWant, order.AuctionID, order.Nonce, order.Signature, hashedOrder)
	if _, err = tx.Exec(insertOrderQuery); err != nil {
		err = fmt.Errorf("Error getting orders from db for placeauctionorder: %s", err)
		return
	}

	logging.Infof("Placed order with id %x!", hashedOrder)

	return
}

// ViewAuctionOrderBook takes in a trading pair and auction ID, and returns auction orders.
func (db *DB) ViewAuctionOrderBook(tradingPair *match.Pair, auctionID [32]byte) (orderbook map[float64][]*match.AuctionOrderIDPair, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for NewAuction: %s", err)
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

	if orderbook, err = db.ViewAuctionOrderBookTx(auctionID, tradingPair, tx); err != nil {
		err = fmt.Errorf("Error viewing auction orderbook for tx: %s", err)
		return
	}

	return
}

func (db *DB) ViewAuctionOrderBookTx(auctionID [32]byte, tradingPair *match.Pair, tx *sql.Tx) (orderbook map[float64][]*match.AuctionOrderIDPair, err error) {

	orderbook = make(map[float64][]*match.AuctionOrderIDPair)
	if _, err = tx.Exec("USE " + db.auctionSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for viewauctionorderbook: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce, sig, hashedOrder FROM %s WHERE auctionID = '%x';", tradingPair, auctionID)
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting orders from db for viewauctionorderbook: %s", err)
		return
	}

	defer func() {
		// TODO: if there's a better way to chain all these errors, figure it out
		var newErr error
		if newErr = rows.Close(); newErr != nil {
			err = fmt.Errorf("Error closing rows for viewauctionorderbook: %s", newErr)
			return
		}
		return
	}()

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.AuctionOrder
	var thisOrderPair *match.AuctionOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var auctionIDBytes []byte
	var nonceBytes []byte
	var sigBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64

	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.AuctionOrder)
		thisOrderPair = new(match.AuctionOrderIDPair)
		if err = rows.Scan(&pkBytes, &thisOrder.Side, &thisPrice, &thisOrder.AmountHave, &thisOrder.AmountWant, &auctionIDBytes, &nonceBytes, &sigBytes, &hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error scanning into order for viewauctionorderbook: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		for _, byteArrayPtr := range []*[]byte{&pkBytes, &auctionIDBytes, &nonceBytes, &sigBytes, &hashedOrderBytes} {
			if *byteArrayPtr, err = hex.DecodeString(string(*byteArrayPtr)); err != nil {
				err = fmt.Errorf("Error decoding bytes for viewauctionorderbook: %s", err)
				return
			}
		}

		// Copy all of the bytes
		copy(thisOrder.Pubkey[:], pkBytes)
		copy(thisOrder.AuctionID[:], auctionIDBytes)
		copy(thisOrder.Signature[:], sigBytes)
		copy(thisOrder.Nonce[:], nonceBytes)
		copy(thisOrderPair.OrderID[:], hashedOrderBytes)

		orderbook[thisPrice] = append(orderbook[thisPrice], thisOrderPair)
		thisOrderPair.Order = thisOrder

	}

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
		if encodedOrder, err = hex.DecodeString(string(encodedOrder)); err != nil {
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

// MatchAuction calculates a single clearing price to execute orders at, and executes at that price.
// TODO: remove "height" from the return
func (db *DB) MatchAuction(auctionID [32]byte) (height uint64, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for MatchAuction: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while matching auction: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Do this for every pair we have

	// Run the matching algorithm for every pair in the orderbook
	for _, pair := range db.pairsArray {
		var orderExecs []*match.OrderExecution
		var settlementExecs []*match.SettlementExecution
		if orderExecs, settlementExecs, err = db.clearingMatchingAlgorithmTx(auctionID, pair, tx); err != nil {
			err = fmt.Errorf("Error running clearing matching algorithm for pair %s: %s", pair, err)
			return
		}
		// now process all of these matches based on the matching algorithm
		for _, exec := range execs {
			if err = db.ProcessExecution(exec, pair, tx); err != nil {
				err = fmt.Errorf("Error processing a single execution for clearing matching algorithm: %s", err)
				return
			}
		}
	}

	return
}

// ProcessOrderExecution handles either deleting or updating a single order that has been executed, depending on whether or not
// it has been filled. It returns a pubkey for the order, and an error. We return a pubkey because that's usually how settlement
// systems settle things, a pubkey is an identifier. It is not needed in matching but it is needed to settle, so that is an output.
func (db *DB) ProcessOrderExecution(exec *match.OrderExecution, pair *match.Pair, tx *sql.Tx) (orderPubkeyBytes []byte, err error) {
	// First use the auction schema
	if _, err = tx.Exec("USE " + db.auctionSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema to process order execution: %s", err)
		return
	}

	// Get the pubkey before we do anything to the order
	var rows *sql.Rows
	getPubkeyQuery := fmt.Sprintf("SELECT pubkey FROM %s WHERE hashedOrder='%x';", pair, exec.OrderID)
	if rows, err = tx.Query(getPubkeyQuery); err != nil {
		err = fmt.Errorf("Error getting pubkey from hashed order in db for process order execution: %s", err)
		return
	}

	// init the pubkey bytes
	if rows.Next() {
		if err = rows.Scan(&orderPubkeyBytes); err != nil {
			err = fmt.Errorf("Error scanning pubkeyBytes from rows for process order execution: %s", err)
			return
		}
	} else {
		err = fmt.Errorf("Could not find pubkeyBytes / order for that order ID")
		return
	}

	// Now close the rows because you sort have to do that
	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for process order execution: %s", err)
		return
	}
	// decode it because mysql is very frustrating
	if orderPubkeyBytes, err = hex.DecodeString(string(orderPubkeyBytes)); err != nil {
		err = fmt.Errorf("Error decoding orderPubkeyBytes from hex for process order executon: %s", err)
		return
	}

	// If the order was filled then delete it. If not then update it.
	if exec.Filled {
		// If the order was filled, delete it from the orderbook
		deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE hashedOrder='%x';", pair.String(), exec.OrderID)
		var res sql.Result
		if res, err = tx.Exec(deleteOrderQuery); err != nil {
			err = fmt.Errorf("Error deleting order within tx for processorderexecution: %s", err)
			return
		}
		// Now we check that there was only one row deleted. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		var rowsAffected int64
		if rowsAffected, err = res.RowsAffected(); err != nil {
			err = fmt.Errorf("Error while getting rows affected for process order execution: %s", err)
			return
		}
		if rowsAffected != 1 {
			logging.Errorf("Error: Order delete should only have affected one row. Instead, it affected %d", rowsAffected)
		}
	} else {
		// If the order was not filled, just update the amounts
		updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE hashedOrder='%x';", pair.String(), exec.NewAmountHave, exec.NewAmountWant, exec.OrderID)
		var res sql.Result
		if res, err = tx.Exec(updateOrderQuery); err != nil {
			err = fmt.Errorf("Error updating order within tx for processorderexecution: %s", err)
			return
		}

		// Now we check that there was only one row updated. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		var rowsAffected int64
		if rowsAffected, err = res.RowsAffected(); err != nil {
			err = fmt.Errorf("Error while getting rows affected for process order execution: %s", err)
			return
		}
		if rowsAffected != 1 {
			logging.Errorf("Error: Order update should only have affected one row. Instead, it affected %d", rowsAffected)
		}
	}
	// TODO
	return
}

// TODO: this part gets removed in the future, done somewhere else. There should be a settlement engine aside from the matching
// engine. Also TODO: Lock funds while in orders and do "swaps" afterwards, crediting as well as debiting when an order is matched.
// ProcessExecutionSettlement processes the settlement part of an order execution. For example, this changes values in a database.
func (db *DB) ProcessExecutionSettlement(exec *match.OrderExecution, orderPubkeyBytes []byte, tx *sql.Tx) (err error) {
	// use the balance schema because we're ending with balance transactions
	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		err = fmt.Errorf("Error using balance schema to process exec settlement: %s", err)
		return
	}

	// Do a bunch of parsing and grabbing pubkeys and coinparams and stuff
	var orderPubkey *koblitz.PublicKey
	if orderPubkey, err = koblitz.ParsePubKey(orderPubkeyBytes, koblitz.S256()); err != nil {
		err = fmt.Errorf("Error parsing pubkeybytes for process execution settlement: %s", err)
		return
	}

	var execCoinType *coinparam.Params
	if execCoinType, err = exec.Debited.Asset.CoinParamFromAsset(); err != nil {
		err = fmt.Errorf("Error getting coin param from debited asset while processing exec settlement: %s", err)
		return
	}

	// Debit the pubkey associated with the order with the amount that it executed for
	if err = db.AddToBalanceWithinTransaction(orderPubkey, exec.Debited.Amount, tx, execCoinType); err != nil {
		err = fmt.Errorf("Error adding to buyorder pubkey balance for fill: %s", err)
		return
	}
	return
}

// ProcessExecution processes an order execution on both the orderbook and settlement side for a single pair. I expect this to
// be moved upwards to the server logic side TODO
func (db *DB) ProcessExecution(exec *match.OrderExecution, pair *match.Pair, tx *sql.Tx) (err error) {

	// First process orders for execution
	// TODO: during the matching / settlement split-up, this should not be called. ProcessOrderExecution should be instead,
	// and the "call settlement" should be a call to the settlement engine. This entire method should instead be part of the
	// OpencxServer logic.
	var orderPubkeyBytes []byte
	if orderPubkeyBytes, err = db.ProcessOrderExecution(exec, pair, tx); err != nil {
		err = fmt.Errorf("Error processing order execution while processing execution: %s", err)
		return
	}

	logging.Warnf("Warning! Settlement not implemented for clearing price matching algorithm!")
	logging.Warnf("The pubkey would have been %x", orderPubkeyBytes)
	// Then call settlement
	// if err = db.ProcessExecutionSettlement(exec, orderPubkeyBytes, tx); err != nil {
	// 	err = fmt.Errorf("Error processing execution settlement while processing execution: %s", err)
	// 	return
	// }
	// TODO

	return
}

// clearingMatchingAlgorithmTx runs the matching algorithm based on clearing price for a single batch pair
func (db *DB) clearingMatchingAlgorithmTx(auctionID [32]byte, pair *match.Pair, tx *sql.Tx) (orderExecs []*match.OrderExecution, settlementExecs []*match.SettlementExecution, err error) {

	// map representation of orderbook
	var book map[float64][]*match.AuctionOrderIDPair
	if book, err = db.ViewAuctionOrderBookTx(auctionID, pair, tx); err != nil {
		err = fmt.Errorf("Error viewing orderbook tx for clearing matching algorithm tx: %s", err)
		return
	}

	// We can now calculate a clearing price and run the matching algorithm
	if orderExecs, settlementExecs, err = match.MatchClearingAlgorithm(book); err != nil {
		err = fmt.Errorf("Error running clearing matching algorithm for match auction: %s", err)
		return
	}

	return
}

/*
TODO: implement this
** This is a note on a matching algorithm that may be more fair than a single clearing price for batch auctions. **
** This also prevents price slippage **

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
This is an example of how we iteratively match orders pro-rata for intersecting prices.

We look at entire prices when doing the calculation for matching, comparing what is provided to what is given.
Iteratively matching pro-rata for *intersecting prices* (ordering: low w/h for sell, high h/w for buy), we add the left over coins from trades into a reserve.

If there are orders with "intersecting" prices, so an order fills another exactly (in which case no coins get added to a reserve), or they are mutually competitive (in which case coins ARE added),
there will be coins left over in a reserve.

If, after all intersecting matches, there are no other orders, then we need to figure out a redistribution strategy, since we have a bunch of coins and no orders to match with them.

If, after all intersecting matches, there are other orders, we will either have reserves in 0 buckets, one bucket, or both.

In the case of zero buckets, we are thankfully done!

In the case of one bucket, we will determine which of the sides would be matched if they were to be given this asset. We then match orders according to their price priority, pro-rata, and then alternate to the other bucket.
This will fill the other bucket, and we continue doing this until there are no coins left in either bucket, or there are no orders left to match (or both!).

In the case of two buckets, we do as described before, starting with the bucket with more funds in it.
The bucket that we start with doesn't matter (at least I don't think).
We use all of these excess funds to match as many orders as possible.
Each pro-rata match will also fill the bucket even more, but decreasingly, since the priority (ordering) defined by price means the orders being matched will get more and more "expensive," and take up more of their
bucket while providing less to the other.

The one & two bucket strategies are actually equivalent.

If we have no more funds then we're done!!!

If, AFTER ALL THIS, we end up with all orders matched, and STILL have funds left in a bucket, we have to default to some sort of redistribution strategy.
But then we're finally done!

If we could generalize this matching algorithm so we don't have to be so iterative and dependent on cases that would be great, but for now this is what we have.

**As an addendum**, if we have either orders left or funds, we could bring this matching algorithm to a whole other level by running the same thing, but drop everything down to a stateful level!
We do the same thing as above, but each auction has some orderbook state. We match like this until completion in the current auction batch of orders.
Next, if we have funds we run out bucket-emptying algorithm, except for with price and time priority, and ties with price AND time priority get settled with pro-rata matching.
If we have orders and no funds, we drop the orders onto the orderbook and match with price/time priority, settling ties with pro-rata matching.

**Note:**
A redistribution strategy just means an answer to "What are we going to do with all of these funds?". It could mean the exchange takes them, or split equally among the matched, or split pro-rata among the matched, who knows.

*/
