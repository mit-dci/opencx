package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// SQLLimitOrderbook is the representation of a limit orderbook for SQL
type SQLLimitOrderbook struct {
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// orderbook schema name
	orderSchema string

	// this pair
	pair *match.Pair
}

// The schema for the limit orderbook
const (
	limitOrderbookSchema = "pubkey VARBINARY(66), orderID VARBINARY(64), side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"
)

// CreateLimitOrderbook creates a limit orderbook based on a pair
func CreateLimitOrderbook(pair *match.Pair) (book match.LimitOrderbook, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateLimitEngine: %s", err)
		return
	}

	// Set values for limit engine
	lo := &SQLLimitOrderbook{
		dbUsername:  conf.DBUsername,
		dbPassword:  conf.DBPassword,
		orderSchema: conf.ReadOnlyOrderSchemaName,
		dbAddr:      addr,
		pair:        pair,
	}

	if err = lo.setupLimitOrderbookTables(); err != nil {
		err = fmt.Errorf("Error setting up limit orderbook tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", lo.dbUsername, lo.dbPassword, lo.dbAddr.Network(), lo.dbAddr.String())
	if lo.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateLimitEngine: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = lo.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	return
}

// setupLimitOrderbookTables sets up the tables needed for the limit orderbook.
// This assumes everything else is set
func (lo *SQLLimitOrderbook) setupLimitOrderbookTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", lo.dbUsername, lo.dbPassword, lo.dbAddr.Network(), lo.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup limit tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup limit tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while matching setup limit tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup limit order tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", lo.orderSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", lo.pair.String(), limitEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating limit orderbook table: %s", err)
		return
	}
	return
}

// UpdateBookExec takes in an order execution and updates the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookExec(orderExec *match.OrderExecution) (err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// UpdateBookCancel takes in an order cancellation and updates the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookCancel(cancel *match.CancelledOrder) (err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// UpdateBookPlace takes in an order, ID, timestamp, and adds the order to the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookPlace(limitIDPair *match.LimitOrderIDPair) (err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetOrder gets an order from an OrderID
func (lo *SQLLimitOrderbook) GetOrder(orderID *match.OrderID) (limOrder *match.LimitOrderIDPair, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// CalculatePrice takes in a pair and returns the calculated price based on the orderbook.
func (lo *SQLLimitOrderbook) CalculatePrice() (price float64, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetOrdersForPubkey gets orders for a specific pubkey.
func (lo *SQLLimitOrderbook) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.LimitOrderIDPair, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// ViewLimitOrderbook takes in a trading pair and returns the orderbook as a map
func (lo *SQLLimitOrderbook) ViewLimitOrderBook() (book map[float64][]*match.LimitOrderIDPair, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// CreateLimitOrderbookMap creates a map of pair to deposit store, given a list of pairs.
func CreateLimitOrderbookMap(pairList []*match.Pair) (orderbookMap map[match.Pair]match.LimitOrderbook, err error) {

	orderbookMap = make(map[match.Pair]match.LimitOrderbook)
	var curLimOrderbook match.LimitOrderbook
	for _, pair := range pairList {
		if curLimOrderbook, err = CreateLimitOrderbook(pair); err != nil {
			err = fmt.Errorf("Error creating single limit engine while creating limit engine map: %s", err)
			return
		}
		orderbookMap[*pair] = curLimOrderbook
	}

	return
}
