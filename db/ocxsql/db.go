package ocxsql

import (
	"database/sql"
	"fmt"
	"log"

	// mysql is just the driver, always interact with database/sql api
	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/match"
	"github.com/mit-dci/opencx/util"
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

// DB contains the sql DB type as well as a logger
type DB struct {
	Keychain             *util.Keychain
	DBHandler            *sql.DB
	logger               *log.Logger
	balanceSchema        string
	depositSchema        string
	pendingDepositSchema string
	orderSchema          string
	assetArray           []string
	pairsArray           []match.Pair
}

// SetupClient sets up the mysql client and driver
func (db *DB) SetupClient() error {
	var err error

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
	err = db.InitializeTables(db.balanceSchema, "name TEXT, balance BIGINT(64)")
	if err != nil {
		return fmt.Errorf("Could not initialize balance tables: \n%s", err)
	}

	// Initialize Deposit tables (order tables soon)
	err = db.InitializeTables(db.depositSchema, "name TEXT, address TEXT")
	if err != nil {
		return fmt.Errorf("Could not initialize deposit tables: \n%s", err)
	}

	// Initialize pending_deposits table
	err = db.InitializeTables(db.pendingDepositSchema, "name TEXT, expectedConfirmHeight INT(32), depositHeight INT(32), amount BIGINT(64), txid TEXT")
	if err != nil {
		return fmt.Errorf("Could not initialize pending deposit tables: \n%s", err)
	}

	// Initialize order table
	// You can have a price up to 8 digits on the left, and 4 on the right of the decimal
	err = db.InitializePairTables(db.orderSchema, "name TEXT, orderID TEXT, side TEXT, price DOUBLE(16,16) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP")
	if err != nil {
		return fmt.Errorf("Could not initialize order tables: \n%s", err)
	}
	return nil
}

// InitializeTables initializes all of the tables necessary for the exchange to run. The schema string can be either balanceSchema or depositSchema.
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
