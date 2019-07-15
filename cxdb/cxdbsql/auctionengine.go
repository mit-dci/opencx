package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// SQLAuctionEngine is the representation of a matching engine for SQL
type SQLAuctionEngine struct {
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// auction orderbook schema name
	auctionOrderSchema string

	// this pair
	pair *match.Pair
}

// The schema for the auction orderbook
const (
	auctionEngineSchema = "pubkey VARBINARY(66), side TEXT, price DOUBLE(30, 2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), auctionID VARBINARY(64), nonce VARBINARY(4), sig BLOB, hashedOrder VARBINARY(64), PRIMARY KEY (hashedOrder)"
)

// CreateAuctionEngineWithConf creates an auction engine, sets up the connection and tables, and returns the auctionengine interface.
func CreateAuctionEngineWithConf(pair *match.Pair, conf *dbsqlConfig) (engine match.AuctionEngine, err error) {

	var ae *SQLAuctionEngine
	if ae, err = CreateAucEngineStructWithConf(pair, conf); err != nil {
		err = fmt.Errorf("Error creating auction engine struct w/ conf for CreateAuctionEngineWithConf: %s", err)
		return
	}

	// now we actually set the return, all checks have passed
	engine = ae
	return
}

// CreateAucEngineStructWithConf creates an auction engine but instead of returning an interface, it returns a struct.
func CreateAucEngineStructWithConf(pair *match.Pair, conf *dbsqlConfig) (engine *SQLAuctionEngine, err error) {
	// Set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for createAuctionEngine: %s", err)
		return
	}

	// Set values
	ae := &SQLAuctionEngine{
		dbUsername:         conf.DBUsername,
		dbPassword:         conf.DBPassword,
		auctionOrderSchema: conf.AuctionSchemaName,
		dbAddr:             addr,
		pair:               pair,
	}

	if err = ae.setupAuctionOrderbookTables(); err != nil {
		err = fmt.Errorf("Error setting up auction orderbook tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", ae.dbUsername, ae.dbPassword, ae.dbAddr.Network(), ae.dbAddr.String())
	if ae.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for createAuctionEngine: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = ae.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// now we actually set the return, all checks have passed
	engine = ae
	return
}

// CreateAuctionEngine creates a SQL Auction Engine as an auction matching engine, with the default conf
func CreateAuctionEngine(pair *match.Pair) (engine match.AuctionEngine, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	if engine, err = CreateAuctionEngineWithConf(pair, conf); err != nil {
		// Breaking err = fmt.Errorf(...) pattern because this is a cleaner way to return errors
		// since the caller in the other case will not be very explicit
		return
	}

	return
}

// setupAuctionOrderbookTables sets up the tables needed for the auction orderbook.
// This assumes the schema name is set
func (ae *SQLAuctionEngine) setupAuctionOrderbookTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", ae.dbUsername, ae.dbPassword, ae.dbAddr.Network(), ae.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup auction tables: %s", err)
		return
	}

	// when we're done close please
	defer rootHandler.Close()

	if err = rootHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// We do this in a transaction because it's more than one operation
	var tx *sql.Tx
	if tx, err = rootHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for setup auction tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while matching setup auction tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup auction order tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", ae.auctionOrderSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", ae.pair.String(), auctionEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating auction orderbook table: %s", err)
		return
	}
	return
}

// DestroyHandler closes the DB handler that we created, and makes it nil
func (ae *SQLAuctionEngine) DestroyHandler() (err error) {
	if ae.DBHandler == nil {
		err = fmt.Errorf("Error, cannot destroy nil handler, please create new engine")
		return
	}
	if err = ae.DBHandler.Close(); err != nil {
		err = fmt.Errorf("Error closing engine handler for DestroyHandler: %s", err)
		return
	}
	ae.DBHandler = nil
	return
}

// PlaceAuctionOrder places an order in the unencrypted datastore. This assumes that the order is valid.
func (ae *SQLAuctionEngine) PlaceAuctionOrder(order *match.AuctionOrder, auctionID *match.AuctionID) (idRes *match.AuctionOrderIDPair, err error) {
	if ae.DBHandler == nil {
		err = fmt.Errorf("Error, cannot place order for nil handler, please create new engine")
		return
	}
	// Do these two things beforehand so we don't have to rollback any tx's

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

	var tx *sql.Tx
	if tx, err = ae.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for PlaceAuctionOrder: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for placeauctionorder: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	if _, err = tx.Exec("USE " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error while placing solved auction order: %s", err)
		return
	}

	logging.Infof("Placing order %s!", order)

	insertOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s', %f, %d, %d, '%x', '%x', '%x', '%x');", ae.pair.String(), order.Pubkey, order.Side, price, order.AmountHave, order.AmountWant, order.AuctionID, order.Nonce, order.Signature, hashedOrder)
	if _, err = tx.Exec(insertOrderQuery); err != nil {
		logging.Errorf("Bad query run: %s", insertOrderQuery)
		err = fmt.Errorf("Error placing order into db for placeauctionorder: %s", err)
		return
	}

	logging.Infof("Placed order with id %x!", hashedOrder)

	// Finally, set the auction order / id pair
	idRes = &match.AuctionOrderIDPair{
		Order: order,
		Price: price,
	}
	copy(idRes.OrderID[:], hashedOrder)

	return
}

// CancelAuctionOrder cancels an auction order, this assumes that the auction order actually exists
func (ae *SQLAuctionEngine) CancelAuctionOrder(orderID *match.OrderID) (cancelled *match.CancelledOrder, cancelSettlement *match.SettlementExecution, err error) {
	if ae.DBHandler == nil {
		err = fmt.Errorf("Error, cannot cancel order with nil handler, please create new engine")
		return
	}

	var tx *sql.Tx
	if tx, err = ae.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for CancelAuctionOrder: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for CancelAuctionOrder: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	if _, err = tx.Exec("USE " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error while cancelling auction order: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, amountHave FROM %s WHERE hashedOrder = '%x';", ae.pair, orderID)
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting order from db for cancelauctionorder: %s", err)
		return
	}

	var actualSide *match.Side

	var pkBytes []byte
	var orderSide string
	var remainingHave uint64
	if rows.Next() {
		// scan the things we can into this order
		if err = rows.Scan(&pkBytes, &orderSide, &remainingHave); err != nil {
			err = fmt.Errorf("Error scanning for order for cancelauctionorder: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		if pkBytes, err = hex.DecodeString(string(pkBytes)); err != nil {
			err = fmt.Errorf("Error decoding pkBytes for cancelauctionorder: %s", err)
			return
		}

		actualSide = new(match.Side)
		if err = actualSide.FromString(orderSide); err != nil {
			err = fmt.Errorf("Error getting side from string for cancelauctionorder: %s", err)
			return
		}

	}

	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE hashedOrder = '%x';", ae.pair.String(), orderID)
	if _, err = tx.Exec(deleteOrderQuery); err != nil {
		err = fmt.Errorf("Error deleting order for cancel auction order: %s", err)
		return
	}

	cancelled = &match.CancelledOrder{
		OrderID: orderID,
	}
	var debitAsset match.Asset
	if *actualSide == match.Buy {
		debitAsset = ae.pair.AssetHave
	} else {
		debitAsset = ae.pair.AssetWant
	}
	cancelSettlement = &match.SettlementExecution{
		Amount: remainingHave,
		Type:   match.Debit,
		Asset:  debitAsset,
	}
	copy(cancelSettlement.Pubkey[:], pkBytes)

	return
}

// MatchAuction calculates a single clearing price to execute orders at, and executes at that price.
func (ae *SQLAuctionEngine) MatchAuctionOrders(auctionID *match.AuctionID) (orderExecs []*match.OrderExecution, settlementExecs []*match.SettlementExecution, err error) {
	if ae.DBHandler == nil {
		err = fmt.Errorf("Error, cannot match orders for nil handler, please create new engine")
		return
	}

	var tx *sql.Tx
	if tx, err = ae.DBHandler.Begin(); err != nil {
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

	// map representation of orderbook
	var book map[float64][]*match.AuctionOrderIDPair
	if book, err = ae.getOrdersTx(auctionID, tx); err != nil {
		err = fmt.Errorf("Error viewing orderbook tx for clearing matching algorithm tx: %s", err)
		return
	}

	// We can now calculate a clearing price and run the matching algorithm
	var newOrderExecs []*match.OrderExecution
	var newSetExecs []*match.SettlementExecution
	if newOrderExecs, newSetExecs, err = match.MatchClearingAlgorithm(book); err != nil {
		err = fmt.Errorf("Error running clearing matching algorithm for match auction: %s", err)
		return
	}

	// now process all of these matches based on the matching algorithm
	if err = ae.processExecutionsTx(newOrderExecs, tx); err != nil {
		err = fmt.Errorf("Error processing a single execution for clearing matching algorithm: %s", err)
		return
	}

	orderExecs = append(orderExecs, newOrderExecs...)
	settlementExecs = append(settlementExecs, newSetExecs...)

	return
}

// getOrdersTx gets all of the orders for the auction ID
func (ae *SQLAuctionEngine) getOrdersTx(auctionID *match.AuctionID, tx *sql.Tx) (orderbook map[float64][]*match.AuctionOrderIDPair, err error) {

	orderbook = make(map[float64][]*match.AuctionOrderIDPair)
	if _, err = tx.Exec("USE " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for viewauctionorderbook: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce, sig, hashedOrder FROM %s WHERE auctionID = '%x';", ae.pair, auctionID)
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
		thisOrder.TradingPair = *ae.pair
		copy(thisOrderPair.OrderID[:], hashedOrderBytes)
		thisOrderPair.Order = thisOrder
		thisOrderPair.Price = thisPrice
		orderbook[thisPrice] = append(orderbook[thisPrice], thisOrderPair)

	}

	return
}

// processExecutionTx processes a batch of order executions
func (ae *SQLAuctionEngine) processExecutionsTx(execs []*match.OrderExecution, tx *sql.Tx) (err error) {
	// First use the auction schema
	if _, err = tx.Exec("USE " + ae.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema to process order execution: %s", err)
		return
	}

	for _, exec := range execs {
		// If the order was filled then delete it. If not then update it.
		if exec.Filled {
			// If the order was filled, delete it from the orderbook
			deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE hashedOrder='%x';", ae.pair.String(), exec.OrderID)
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
			updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE hashedOrder='%x';", ae.pair.String(), exec.NewAmountHave, exec.NewAmountWant, exec.OrderID)
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
	}
	return
}

// CreateAuctionEngineMap creates a map of pair to auction engine, given a list of pairs.
func CreateAuctionEngineMap(pairList []*match.Pair) (aucMap map[match.Pair]match.AuctionEngine, err error) {

	aucMap = make(map[match.Pair]match.AuctionEngine)
	var curAucEng match.AuctionEngine
	for _, pair := range pairList {
		if curAucEng, err = CreateAuctionEngine(pair); err != nil {
			err = fmt.Errorf("Error creating single auction engine while creating auction engine map: %s", err)
			return
		}
		aucMap[*pair] = curAucEng
	}

	return
}
