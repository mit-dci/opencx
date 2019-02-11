package ocxsql

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

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
	rootPass             = ""
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
	logger               *log.Logger
	balanceSchema        string
	depositSchema        string
	pendingDepositSchema string
	orderSchema          string
	assetArray           []string
	pairsArray           []match.Pair
	globalReads          int64
	globalWrites         int64
	gReadsMutex          sync.Mutex
	gWritesMutex         sync.Mutex
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

// IncrementReads increments the read variable and might print it out, this is just for debugging / making sure we're not overloading the DB
func (db *DB) IncrementReads() {
	db.gReadsMutex.Lock()
	db.globalReads++
	if db.globalReads%100 == 0 {
		logging.Infof("Global reads: %d\n", db.globalReads)
	}
	db.gReadsMutex.Unlock()
}

// IncrementWrites increments the write variable and might print it out, this is just for debugging / making sure we're not overloading the DB
func (db *DB) IncrementWrites() {
	db.gWritesMutex.Lock()
	db.globalWrites++
	if db.globalWrites%100 == 0 {
		logging.Infof("Global writes: %d\n", db.globalWrites)
	}
	db.gWritesMutex.Unlock()
}

// SetupClient sets up the mysql client and driver
func (db *DB) SetupClient() error {
	var err error

	db.gPriceMap = make(map[string]float64)
	db.balanceSchema = balanceSchema
	db.depositSchema = depositSchema
	db.pendingDepositSchema = pendingDepositSchema
	db.orderSchema = orderSchema
	// Create users and schemas and assign permissions to opencx
	err = db.RootInitSchemas(rootPass)
	if err != nil {
		return fmt.Errorf("Root could not initialize schemas: \n%s", err)
	}

	// open db handle
	dbHandle, err := sql.Open("mysql", defaultUsername+":"+defaultPassword+"@/")
	if err != nil {
		return fmt.Errorf("Error opening database: \n%s", err)
	}

	db.DBHandler = dbHandle
	db.assetArray = assetArray
	db.pairsArray = match.GenerateAssetPairs()

	err = db.DBHandler.Ping()
	if err != nil {
		return fmt.Errorf("Could not ping the database, is it running: \n%s", err)
	}

	// Initialize Balance tables (order tables soon)
	// hacky workaround to get behind the fact I made a dumb abstraction with InitializeTables
	err = db.InitializeNewTables(db.balanceSchema, "name TEXT, balance BIGINT(64)")
	if err != nil {
		return fmt.Errorf("Could not initialize balance tables: \n%s", err)
	}
	if err = db.InitializeBalancesFromNames(); err != nil {
		return err
	}

	// Initialize Deposit tables (order tables soon)
	err = db.InitializeTables(db.depositSchema, "name TEXT, address TEXT")
	if err != nil {
		return fmt.Errorf("Could not initialize deposit tables: \n%s", err)
	}

	// Initialize pending_deposits table
	err = db.InitializeNewTables(db.pendingDepositSchema, "name TEXT, expectedConfirmHeight INT(32), depositHeight INT(32), amount BIGINT(64), txid TEXT")
	if err != nil {
		return fmt.Errorf("Could not initialize pending deposit tables: \n%s", err)
	}

	// Initialize order table
	// You can have a price up to 30 digits total, and 10 decimal places.
	err = db.InitializePairTables(db.orderSchema, "name TEXT, orderID TEXT, side TEXT, price DOUBLE(30,10) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP")
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

// InitializeBalancesFromNames makes balance stuff from names in deposits
func (db *DB) InitializeBalancesFromNames() error {

	// use the deposit schema for names

	if _, err := db.DBHandler.Exec("USE " + db.depositSchema + ";"); err != nil {
		return fmt.Errorf("Could not use %s schema: \n%s", db.depositSchema, err)
	}

	var nameArray []string
	if len(db.assetArray) > 0 {
		takeNamesQuery := fmt.Sprintf("SELECT name FROM %s;", db.assetArray[0])
		nameRows, err := db.DBHandler.Query(takeNamesQuery)
		if err != nil {
			return fmt.Errorf("An error occurred while trying to take names: \n%s", err)
		}

		for nameRows.Next() {
			var name string
			if err = nameRows.Scan(&name); err != nil {
				return err
			}

			logging.Infof("name: %s\n", name)

			nameArray = append(nameArray, name)
		}
		if err = nameRows.Close(); err != nil {
			return err
		}
	}

	if _, err := db.DBHandler.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return fmt.Errorf("Could not use %s schema: \n%s", db.balanceSchema, err)
	}

	for _, name := range nameArray {
		for _, coinSchema := range db.assetArray {
			// Create the balance
			insertBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", coinSchema, name, 0)
			if _, err := db.DBHandler.Exec(insertBalanceQuery); err != nil {
				return err
			}
		}
	}
	return nil
}

// RootInitSchemas initalizes the schemas, creates users, and grants permissions to those users
func (db *DB) RootInitSchemas(rootPassword string) error {
	var err error

	// Log in to root
	rootHandler, err := sql.Open("mysql", "root:"+rootPassword+"@/")
	if err != nil {
		return fmt.Errorf("Error opening root db: \n%s", err)
	}

	// When the method is done, close the root connection
	defer rootHandler.Close()

	err = rootHandler.Ping()
	if err != nil {
		return fmt.Errorf("Could not ping the database, is it running: \n%s", err)
	}

	createUserQuery := fmt.Sprintf("CREATE OR REPLACE USER '%s'@'localhost' IDENTIFIED BY '%s';", defaultUsername, defaultPassword)
	_, err = rootHandler.Exec(createUserQuery)
	if err != nil {
		return fmt.Errorf("Could not create default user: \n%s", err)
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

	// grant permissions to default user
	schemaQuery := fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, CREATE, DELETE, DROP ON %s.* TO '%s'@'localhost';", schemaString, username)
	_, err = rootHandler.Exec(schemaQuery)
	if err != nil {
		return fmt.Errorf("Could not grant permissions to %s while creating %s schema: \n%s", schemaString, username, err)
	}

	return nil
}
