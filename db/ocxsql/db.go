package ocxsql

import (
	"database/sql"
	"fmt"

	// mysql is just the driver, always interact with database/sql api
	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/match"
)

// turn into config options
var (
	defaultUsername = "opencx"
	defaultPassword = "testpass"

	// definitely move this to a config file
	balanceSchema        = "balances"
	depositSchema        = "deposit"
	pendingDepositSchema = "pending_deposits"
	orderSchema          = "orders"
	assetArray           = []string{"btc", "ltc", "vtc"}
)

// the globalread and globalwrite variables are for debugging

// DB contains the sql DB type as well as a logger.
// The database is a BEHEMOTH, should be refactored. Some examples on how to refactor are cleaning up mutexes, creating config file for all the globals,
// figuring out a better way to handle schemas, finding a better spot for a keychain, and putting the price somewhere else.
// What would be great is to move everything having to do with price and matching into match and making match more like a matching engine framework
// or library for exchanges.
type DB struct {
	DBHandler            *sql.DB
	balanceSchema        string
	depositSchema        string
	pendingDepositSchema string
	orderSchema          string
	assetArray           []match.Asset
	pairsArray           []*match.Pair
	gPriceMap            map[string]float64
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

// SetupClient sets up the mysql client and driver
func (db *DB) SetupClient(assets []match.Asset, pairs []*match.Pair) error {
	var err error

	db.gPriceMap = make(map[string]float64)
	db.balanceSchema = balanceSchema
	db.depositSchema = depositSchema
	db.pendingDepositSchema = pendingDepositSchema
	db.orderSchema = orderSchema
	// Create users and schemas and assign permissions to opencx
	err = db.RootInitSchemas()
	if err != nil {
		return fmt.Errorf("Root could not initialize schemas: \n%s", err)
	}

	// open db handle
	dbHandle, err := sql.Open("mysql", defaultUsername+":"+defaultPassword+"@/")
	if err != nil {
		return fmt.Errorf("Error opening database: \n%s", err)
	}

	db.DBHandler = dbHandle
	db.assetArray = assets
	db.pairsArray = match.GenerateAssetPairs()

	err = db.DBHandler.Ping()
	if err != nil {
		return fmt.Errorf("Could not ping the database, is it running: \n%s", err)
	}

	// Initialize Balance tables
	// hacky workaround to get behind the fact I made a dumb abstraction with InitializeTables
	err = db.InitializeNewTables(db.balanceSchema, "pubkey TEXT, balance BIGINT(64)")
	if err != nil {
		return fmt.Errorf("Could not initialize balance tables: \n%s", err)
	}

	// Initialize Deposit tables
	err = db.InitializeTables(db.depositSchema, "pubkey VARBINARY(33), address VARCHAR(34), CONSTRAINT unique_pubkeys UNIQUE (pubkey, address)")
	if err != nil {
		return fmt.Errorf("Could not initialize deposit tables: \n%s", err)
	}

	// Initialize pending_deposits table
	err = db.InitializeNewTables(db.pendingDepositSchema, "pubkey VARBINARY(33), expectedConfirmHeight INT(32), depositHeight INT(32), amount BIGINT(64), txid TEXT")
	if err != nil {
		return fmt.Errorf("Could not initialize pending deposit tables: \n%s", err)
	}

	// Initialize order table
	// You can have a price up to 30 digits total, and 10 decimal places.
	err = db.InitializePairTables(db.orderSchema, "pubkey VARBINARY(33), orderID TEXT, side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP")
	if err != nil {
		return fmt.Errorf("Could not initialize order tables: \n%s", err)
	}
	return nil
}

// InitializeTables initializes all of the tables necessary for the exchange to run.
func (db *DB) InitializeTables(schemaName string, schemaSpec string) error {
	var err error

	// Use the schema
	_, err = db.DBHandler.Exec("USE " + schemaName + ";")
	if err != nil {
		return fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
	}
	for _, assetString := range db.assetArray {
		tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", assetString, schemaSpec)
		_, err = db.DBHandler.Exec(tableQuery)
		if err != nil {
			return fmt.Errorf("Could not create table %s: \n%s", assetString, err)
		}
	}
	return nil
}

// InitializeNewTables initalizes tables based on schema and clears them.
func (db *DB) InitializeNewTables(schemaName string, schemaSpec string) error {
	var err error

	// Use the schema
	_, err = db.DBHandler.Exec("USE " + schemaName + ";")
	if err != nil {
		return fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
	}
	for _, assetString := range db.assetArray {
		tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", assetString, schemaSpec)
		_, err = db.DBHandler.Exec(tableQuery)
		if err != nil {
			return fmt.Errorf("Could not create table %s: \n%s", assetString, err)
		}
		deleteQuery := fmt.Sprintf("DELETE FROM %s;", assetString)
		_, err = db.DBHandler.Exec(deleteQuery)
		if err != nil {
			return fmt.Errorf("Could not delete stuff from table after creating: \n%s", err)
		}
	}
	return nil
}

// InitializePairTables initializes tables per pair
func (db *DB) InitializePairTables(schemaName string, schemaSpec string) error {
	var err error

	// Use the schema
	_, err = db.DBHandler.Exec("USE " + schemaName + ";")
	if err != nil {
		return fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
	}
	for _, pair := range db.pairsArray {
		tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", pair.String(), schemaSpec)
		_, err = db.DBHandler.Exec(tableQuery)
		if err != nil {
			return fmt.Errorf("Could not create table %s: \n%s", pair.String(), err)
		}
	}
	return nil
}

// RootInitSchemas initalizes the schemas, creates users, and grants permissions to those users
func (db *DB) RootInitSchemas() error {
	var err error

	// Log in to root
	rootHandler, err := sql.Open("mysql", defaultUsername+":"+defaultPassword+"@/")
	if err != nil {
		return fmt.Errorf("Error opening root db: \n%s", err)
	}

	// When the method is done, close the root connection
	defer rootHandler.Close()

	err = rootHandler.Ping()
	if err != nil {
		return fmt.Errorf("Could not ping the database, is it running: \n%s", err)
	}

	err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.balanceSchema)
	if err != nil {
		return fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
	}

	err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.depositSchema)
	if err != nil {
		return fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
	}

	err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.pendingDepositSchema)
	if err != nil {
		return fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
	}

	err = rootCreateSchemaForUser(rootHandler, defaultUsername, db.orderSchema)
	if err != nil {
		return fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
	}

	return nil
}

// Helper function for db
func rootCreateSchemaForUser(rootHandler *sql.DB, username string, schemaString string) error {
	var err error

	// check pending deposit schema
	// if pending deposit schema not there make it
	_, err = rootHandler.Exec("CREATE SCHEMA IF NOT EXISTS " + schemaString + ";")
	if err != nil {
		return fmt.Errorf("Could not create %s schema: \n%s", schemaString, err)
	}

	return nil
}
