package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

// SQLPuzzleStore is a puzzle store representation for a SQL database
type SQLPuzzleStore struct {
	DBHandler *sql.DB

	// db username
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// puzzle schema name
	puzzleSchema string

	// the pair for this puzzle store
	// this is just for convenience, the protocol still works if you have one massive puzzle store
	// but if you run many markets at once then you may want to invalidate orders that weren't submitted
	// for the pair they said they were
	pair *match.Pair
}

const (
	puzzleStoreSchema = "encodedOrder TEXT, auctionID VARBINARY(64), selected BOOLEAN"
)

// CreatePuzzleStore creates a puzzle store for a specific coin.
func CreatePuzzleStore(pair *match.Pair) (store cxdb.PuzzleStore, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// Set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateSQLPuzzleStore: %s", err)
		return
	}

	// Set values
	sp := &SQLPuzzleStore{
		dbUsername:   conf.DBUsername,
		dbPassword:   conf.DBPassword,
		puzzleSchema: conf.PuzzleSchemaName,
		dbAddr:       addr,
		pair:         pair,
	}

	if err = sp.setupPuzzleStoreTables(); err != nil {
		err = fmt.Errorf("Error setting up settlement store tables while creating store: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", sp.dbUsername, sp.dbPassword, sp.dbAddr.Network(), sp.dbAddr.String())
	if sp.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateSQLPuzzleStore: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = sp.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// Now we actually set the engine
	store = sp
	return
}

// ViewAuctionPuzzleBook takes in an auction ID, and returns encrypted auction orders, and puzzles.
// You don't know what auction IDs should be in the orders encrypted in the puzzle book, but this is
// what was submitted.
// This only selects "selected" orders, so while some could be censored and not shown, there is currently
// no functionality for censoring orders and setting the 'selected' flag to false.
func (sp *SQLPuzzleStore) ViewAuctionPuzzleBook(auctionID *match.AuctionID) (puzzles []*match.EncryptedAuctionOrder, err error) {
	// ACID
	var tx *sql.Tx
	if tx, err = sp.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for ViewAuctionPuzzleBook: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for ViewAuctionPuzzleBook: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the order schema
	if _, err = tx.Exec("USE " + sp.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for ViewAuctionPuzzleBook: %s", err)
		return
	}

	var serializedPuzzle []byte
	var rows *sql.Rows
	getPuzzleBookQuery := fmt.Sprintf("SELECT encodedOrder FROM %s WHERE auctionID='%x' AND selected='%t';", sp.pair.String(), auctionID[:], true)
	if rows, err = tx.Query(getPuzzleBookQuery); err != nil {
		err = fmt.Errorf("Error querying for puzzles for ViewAuctionPuzzleBook: %s", err)
		return
	}

	var currPuzzle *match.EncryptedAuctionOrder
	for rows.Next() {
		if err = rows.Scan(&serializedPuzzle); err != nil {
			err = fmt.Errorf("Error scanning puzzle for ViewAuctionPuzzleBook: %s", err)
			return
		}

		if serializedPuzzle, err = hex.DecodeString(string(serializedPuzzle)); err != nil {
			err = fmt.Errorf("Error decoding hex string serializedPuzzle for ViewAuctionPuzzleBook: %s", err)
			return
		}

		// Just deserialize
		currPuzzle = new(match.EncryptedAuctionOrder)
		if err = currPuzzle.Deserialize(serializedPuzzle); err != nil {
			err = fmt.Errorf("Error deserializing current puzzle for ViewAuctionPuzzleBook: %s", err)
			return
		}

	}
	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for ViewAuctionPuzzleBook: %s", err)
		return
	}

	return
}

// PlaceAuctionPuzzle puts an encrypted auction order in the datastore.
func (sp *SQLPuzzleStore) PlaceAuctionPuzzle(puzzledOrder *match.EncryptedAuctionOrder) (err error) {
	// ACID
	var tx *sql.Tx
	if tx, err = sp.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for PlaceAuctionPuzzle: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for PlaceAuctionPuzzle: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the order schema
	if _, err = tx.Exec("USE " + sp.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for PlaceAuctionPuzzle: %s", err)
		return
	}

	var pzOrderBytes []byte
	if pzOrderBytes, err = puzzledOrder.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing puzzled order for PlaceAuctionPuzzle: %s", err)
		return
	}

	defaultSelected := true
	insertPuzzleQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%x', %t);", sp.pair.String(), pzOrderBytes, puzzledOrder.IntendedAuction, defaultSelected)
	if _, err = tx.Exec(insertPuzzleQuery); err != nil {
		err = fmt.Errorf("Error placing puzzle into db for PlaceAuctionPuzzle: %s", err)
		return
	}
	return
}

// setupPuzzleStoreTables sets up the tables needed for the auction orderbook.
// This assumes the schema name is set
func (sp *SQLPuzzleStore) setupPuzzleStoreTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", sp.dbUsername, sp.dbPassword, sp.dbAddr.Network(), sp.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup puzzle store tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup puzzle store tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while creating puzzle store tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + sp.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup puzzle store tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + sp.puzzleSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", sp.puzzleSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", sp.pair.String(), puzzleStoreSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating puzzle store table: %s", err)
		return
	}
	return
}

// CreatePuzzleStoreMap creates a map of pair to pair list, given a list of pairs.
func CreatePuzzleStoreMap(pairList []*match.Pair) (pzMap map[match.Pair]cxdb.PuzzleStore, err error) {

	pzMap = make(map[match.Pair]cxdb.PuzzleStore)
	var curPzEng cxdb.PuzzleStore
	for _, pair := range pairList {
		if curPzEng, err = CreatePuzzleStore(pair); err != nil {
			err = fmt.Errorf("Error creating single puzzle store while creating puzzle store map: %s", err)
			return
		}
		pzMap[*pair] = curPzEng
	}

	return
}
