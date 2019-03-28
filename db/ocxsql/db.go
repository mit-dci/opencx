package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/coinparam"

	// mysql is just the driver, always interact with database/sql api
	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// turn into config options
var (
	defaultUsername = "opencx"
	defaultPassword = "testpass"

	// definitely move this to a config file
	balanceSchema          = "balances"
	depositSchema          = "deposit"
	pendingDepositSchema   = "pending_deposits"
	orderSchema            = "orders"
	lightningbalanceSchema = "litbalances"
)

// the globalread and globalwrite variables are for debugging

// DB contains the sql DB type as well as a logger.
// The database is a BEHEMOTH, should be refactored. Some examples on how to refactor are cleaning up mutexes, creating config file for all the globals,
// What would be great is to move everything having to do with price and matching into match and making match more like a matching engine framework
// or library for exchanges.
type DB struct {
	// the SQL handler for the db
	DBHandler *sql.DB

	// name of balance schema
	balanceSchema string

	// name of deposit schema
	depositSchema string

	// name of pending deposit schema
	pendingDepositSchema string

	// name of order schema
	orderSchema string

	// name of lightning balance schema
	lightningbalanceSchema string

	// list of all coins supported, passed in from above
	coinList []*coinparam.Params

	// the pairs that are supported. generated from coinList when the db is initiated
	pairsArray []*match.Pair

	// pricemap for pair that we manually add to
	gPriceMap map[string]float64
}

// SetPrice sets the price, uses a lock since it will be written to and read from possibly at the same time (written to by server, read by client)
func (db *DB) SetPrice(newPrice float64, pairString string) {
	db.gPriceMap[pairString] = newPrice
}

// GetPrice returns the price and side of the last transacted price
func (db *DB) GetPrice(pairString string) (price float64) {
	price, found := db.gPriceMap[pairString]
	if !found {
		return float64(0)
	}
	return price
}

// GetPairs returns the pairs list
func (db *DB) GetPairs() (pairArray []*match.Pair) {
	pairArray = db.pairsArray
	return
}

// SetupClient sets up the mysql client and driver
func (db *DB) SetupClient(coinList []*coinparam.Params) (err error) {
	db.gPriceMap = make(map[string]float64)
	db.balanceSchema = balanceSchema
	db.depositSchema = depositSchema
	db.pendingDepositSchema = pendingDepositSchema
	db.orderSchema = orderSchema
	db.lightningbalanceSchema = lightningbalanceSchema
	// Create users and schemas and assign permissions to opencx
	if err = db.RootInitSchemas(); err != nil {
		err = fmt.Errorf("Root could not initialize schemas: \n%s", err)
		return
	}

	// open db handle
	dbHandle, err := sql.Open("mysql", defaultUsername+":"+defaultPassword+"@/")
	if err != nil {
		err = fmt.Errorf("Error opening database: \n%s", err)
		return
	}

	db.DBHandler = dbHandle
	db.coinList = coinList

	// Get all the pairs
	if db.pairsArray, err = match.GenerateAssetPairs(coinList); err != nil {
		return
	}

	// DEBUGGING
	// Get all the assets
	for i, asset := range db.coinList {
		logging.Debugf("Asset %d: %s\n", i, asset.Name)
	}

	// Get all the asset pairs
	for i, pair := range db.pairsArray {
		logging.Debugf("Pair %d: %s\n", i, pair)
	}
	// END DEBUGGING

	err = db.DBHandler.Ping()
	if err != nil {
		return fmt.Errorf("Could not ping the database, is it running: \n%s", err)
	}

	// Initialize Balance tables
	// hacky workaround to get behind the fact I made a dumb abstraction with InitializeTables
	// 66 bytes because we use big bytes and they use small bytes for varbinary
	if err = db.InitializeNewTables(db.balanceSchema, "pubkey VARBINARY(66), balance BIGINT(64)"); err != nil {
		err = fmt.Errorf("Could not initialize balance tables: \n%s", err)
		return
	}

	// Initialize Deposit tables
	if err = db.InitializeTables(db.depositSchema, "pubkey VARBINARY(66), address VARCHAR(34), CONSTRAINT unique_pubkeys UNIQUE (pubkey, address)"); err != nil {
		err = fmt.Errorf("Could not initialize deposit tables: \n%s", err)
		return
	}

	// Initialize pending_deposits table
	if err = db.InitializeNewTables(db.pendingDepositSchema, "pubkey VARBINARY(66), expectedConfirmHeight INT(32), depositHeight INT(32), amount BIGINT(64), txid TEXT"); err != nil {
		err = fmt.Errorf("Could not initialize pending deposit tables: \n%s", err)
		return
	}

	// Initialize order table
	// You can have a price up to 30 digits total, and 10 decimal places.
	if err = db.InitializePairTables(db.orderSchema, "pubkey VARBINARY(66), orderID TEXT, side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"); err != nil {
		err = fmt.Errorf("Could not initialize order tables: \n%s", err)
		return
	}

	if err = db.InitializeNewTables(db.lightningbalanceSchema, "pubkey VARBINARY(66), qchanID INT(32), amount BIGINT(64)"); err != nil {
		err = fmt.Errorf("Could not initialize lightning balance tables: \n%s", err)
		return
	}
	return
}

// InitializeTables initializes all of the tables necessary for the exchange to run.
func (db *DB) InitializeTables(schemaName string, schemaSpec string) (err error) {

	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}
	for _, chain := range db.coinList {
		tableQuery := fmt.Sprintf("CREATE OR REPLACE TABLE %s (%s);", chain.Name, schemaSpec)
		if _, err = db.DBHandler.Exec(tableQuery); err != nil {
			err = fmt.Errorf("Could not create table %s: \n%s", chain.Name, err)
			return
		}
	}
	return
}

// InitializeNewTables initalizes tables based on schema and clears them.
func (db *DB) InitializeNewTables(schemaName string, schemaSpec string) (err error) {
	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}
	for _, chain := range db.coinList {
		tableQuery := fmt.Sprintf("CREATE OR REPLACE TABLE %s (%s);", chain.Name, schemaSpec)
		if _, err = db.DBHandler.Exec(tableQuery); err != nil {
			err = fmt.Errorf("Could not create table %s: \n%s", chain.Name, err)
			return
		}
		deleteQuery := fmt.Sprintf("DELETE FROM %s;", chain.Name)
		if _, err = db.DBHandler.Exec(deleteQuery); err != nil {
			err = fmt.Errorf("Could not delete stuff from table after creating: \n%s", err)
			return
		}
	}
	return
}

// InitializePairTables initializes tables per pair
func (db *DB) InitializePairTables(schemaName string, schemaSpec string) (err error) {
	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}
	for _, pair := range db.pairsArray {
		tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", pair.String(), schemaSpec)
		if _, err = db.DBHandler.Exec(tableQuery); err != nil {
			err = fmt.Errorf("Could not create table %s: \n%s", pair.String(), err)
			return
		}
	}
	return
}

// RootInitSchemas initalizes the schemas, creates users, and grants permissions to those users
func (db *DB) RootInitSchemas() (err error) {
	// Log in to root
	rootHandler, err := sql.Open("mysql", defaultUsername+":"+defaultPassword+"@/")
	if err != nil {
		return fmt.Errorf("Error opening root db: \n%s", err)
	}

	// When the method is done, close the root connection
	defer rootHandler.Close()

	if err = rootHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: \n%s", err)
		return
	}

	if err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.balanceSchema); err != nil {
		err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
		return
	}

	if err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.depositSchema); err != nil {
		err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
		return
	}

	if err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.pendingDepositSchema); err != nil {
		err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
		return
	}

	if err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.orderSchema); err != nil {
		err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
		return
	}

	if err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.lightningbalanceSchema); err != nil {
		err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
		return
	}

	return
}

// Helper function for db
func rootCreateSchemaForUser(rootHandler *sql.DB, username string, schemaString string) (err error) {
	// check pending deposit schema
	// if pending deposit schema not there make it
	if _, err = rootHandler.Exec("CREATE SCHEMA IF NOT EXISTS " + schemaString + ";"); err != nil {
		err = fmt.Errorf("Could not create %s schema: \n%s", schemaString, err)
		return
	}

	return
}
