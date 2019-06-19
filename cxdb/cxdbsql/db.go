package cxdbsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/mit-dci/lit/coinparam"

	// mysql is just the driver, always interact with database/sql api
	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/opencx/match"
)

type dbsqlConfig struct {
	// Filename of config file where this stuff can be set as well
	ConfigFile string

	// database home dir
	DBHomeDir string `long:"dir" description:"Location of the root directory for the sql db info and config"`

	// database info required to establish connection
	DBUsername string `long:"dbuser" description:"database username"`
	DBPassword string `long:"dbpassword" description:"database password"`
	DBHost     string `long:"dbhost" description:"Host for the database connection"`
	DBPort     uint16 `long:"dbport" description:"Port for the database connection"`

	// database schema names
	ReadOnlyBalanceSchemaName string `long:"readonlybalanceschema" description:"Name of read-only balance schema"`
	BalanceSchemaName         string `long:"balanceschema" description:"Name of balance schema"`
	DepositSchemaName         string `long:"depositschema" description:"Name of deposit schema"`
	PendingDepositSchemaName  string `long:"penddepschema" description:"Name of pending deposit schema"`
	PuzzleSchemaName          string `long:"puzzleschema" description:"Name of schema for puzzle orderbooks"`
	AuctionSchemaName         string `long:"auctionschema" description:"Name of schema for auction ID"`
	AuctionOrderSchemaName    string `long:"auctionorderschema" description:"Name of schema for auction orderbook"`
	OrderSchemaName           string `long:"orderschema" description:"Name of schema for limit orderbook"`
	PeerSchemaName            string `long:"peerschema" description:"Name of schema for peer storage"`

	// database table names
	PuzzleTableName       string `long:"puzzletable" description:"Name of table for puzzle orderbooks"`
	AuctionOrderTableName string `long:"auctionordertable" description:"Name of table for auction orders"`
	PeerTableName         string `long:"peertable" description:"Name of table for peer storage"`
}

// Let these be turned into config things at some point
var (
	defaultConfigFilename = "sqldb.conf"
	defaultHomeDir        = os.Getenv("HOME")
	defaultDBHomeDirName  = defaultHomeDir + "/.opencx/db/"
	defaultDBPort         = uint16(3306)
	defaultDBHost         = "localhost"
	defaultDBUser         = "opencx"
	defaultDBPass         = "testpass"

	// definitely move this to a config file
	defaultReadOnlyBalanceSchema = "balances_readonly"
	defaultBalanceSchema         = "balances"
	defaultDepositSchema         = "deposit"
	defaultPendingDepositSchema  = "pending_deposits"
	defaultPuzzleSchema          = "puzzle"
	defaultAuctionSchema         = "auctions"
	defaultAuctionOrderSchema    = "auctionorder"
	defaultOrderSchema           = "orders"
	defaultPeerSchema            = "peers"

	// tables
	defaultAuctionOrderTable = "auctionorders"
	defaultPuzzleTable       = "puzzles"
	defaultPeerTable         = "opencxpeers"

	// Set defaults
	defaultConf = &dbsqlConfig{
		// home dir
		DBHomeDir: defaultDBHomeDirName,

		// user / pass / net stuff
		DBUsername: defaultDBUser,
		DBPassword: defaultDBPass,
		DBHost:     defaultDBHost,
		DBPort:     defaultDBPort,

		// schemas
		ReadOnlyBalanceSchemaName: defaultReadOnlyBalanceSchema,
		BalanceSchemaName:         defaultBalanceSchema,
		DepositSchemaName:         defaultDepositSchema,
		PendingDepositSchemaName:  defaultPendingDepositSchema,
		PuzzleSchemaName:          defaultPuzzleSchema,
		AuctionSchemaName:         defaultAuctionSchema,
		AuctionOrderSchemaName:    defaultAuctionOrderSchema,
		OrderSchemaName:           defaultOrderSchema,
		PeerSchemaName:            defaultPeerSchema,

		// tables
		PuzzleTableName:       defaultPuzzleTable,
		AuctionOrderTableName: defaultAuctionOrderTable,
		PeerTableName:         defaultPeerTable,
	}
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *dbsqlConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

// TODO (rjected): Before any abstractions are done or ANYTHING except build new features in the same way we've already built them, we NEED to give this db
// either a config file or something.

// DB contains the sql DB type as well as a logger.
// The database is a BEHEMOTH, should be refactored. Some examples on how to refactor are cleaning up mutexes, creating config file for all the globals,
// What would be great is to move everything having to do with price and matching into match and making match more like a matching engine framework
// or library for exchanges. This should conform to the cxdb interface, and if the server uses the noise protocol / authenticated networking, or anything
// that requires conforming to the lncore.LitPeerStorage interface, it should conform to that as well.
type DB struct {
	// the SQL handler for the db
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string
	// db host and port
	dbAddr net.Addr

	// standard exchange stuff
	// name of balance schema
	balanceSchema string
	// name of deposit schema
	depositSchema string
	// name of pending deposit schema
	pendingDepositSchema string
	// name of order schema
	orderSchema string

	// peer schema stuff
	// name of peer schema
	peerSchema string
	// name of peer table
	peerTableName string

	// auction schema stuff
	// name of puzzle orderbook schema
	puzzleSchema string
	// name of the puzzle orderbook table
	puzzleTable string
	// name of the auction orderbook schema
	auctionSchema string
	// name of the (auction ID => auction number) map schema
	auctionOrderSchema string
	// name of the (auction ID => auction number) table
	auctionOrderTable string

	// list of all coins supported, passed in from above
	coinList []*coinparam.Params

	// the pairs that are supported. generated from coinList when the db is initiated
	pairsArray []*match.Pair

	// pricemap for pair that we manually add to
	gPriceMap map[string]float64
	// priceMapMtx is a lock for gPriceMap
	priceMapMtx *sync.Mutex
}

// SetPrice sets the price, uses a lock since it will be written to and read from possibly at the same time (written to by server, read by client)
func (db *DB) SetPrice(newPrice float64, pairString string) {
	db.priceMapMtx.Lock()
	db.gPriceMap[pairString] = newPrice
	db.priceMapMtx.Unlock()
}

// GetPrice returns the price and side of the last transacted price
func (db *DB) GetPrice(pairString string) (price float64, err error) {
	var found bool
	if price, found = db.gPriceMap[pairString]; !found {
		err = fmt.Errorf("Could not get price, pair not found")
	}
	return
}

// GetPairs returns the pairs list
func (db *DB) GetPairs() (pairArray []*match.Pair) {
	pairArray = db.pairsArray
	return
}

func (db *DB) setOptionsFromConfig(conf *dbsqlConfig) (err error) {

	// create and parse config file
	dbConfigSetup(conf)

	// username & pass stuff
	db.dbUsername = conf.DBUsername
	db.dbPassword = conf.DBPassword

	// schemas
	db.balanceSchema = conf.BalanceSchemaName
	db.depositSchema = conf.DepositSchemaName
	db.pendingDepositSchema = conf.PendingDepositSchemaName
	db.orderSchema = conf.OrderSchemaName
	db.peerSchema = conf.PeerSchemaName
	db.puzzleSchema = conf.PuzzleSchemaName
	db.auctionSchema = conf.AuctionSchemaName
	db.auctionOrderSchema = conf.AuctionOrderSchemaName

	// tables
	db.puzzleTable = conf.PuzzleTableName
	db.auctionOrderTable = conf.AuctionOrderTableName
	db.peerTableName = conf.PeerTableName

	// Resolve address for host and port, then set that as the network address
	if db.dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Error resolving database address: \n%s", err)
		return
	}

	return
}

// SetupClient sets up the mysql client and driver
func (db *DB) SetupClient(coinList []*coinparam.Params) (err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// We created this config, now we set the options from the config
	if err = db.setOptionsFromConfig(defaultConf); err != nil {
		err = fmt.Errorf("Error setting options from config for setupclient: %s", err)
		return
	}

	db.gPriceMap = make(map[string]float64)
	db.priceMapMtx = new(sync.Mutex)
	db.coinList = coinList

	// Resolve address for host and port, then set that as the network address
	if db.dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Error resolving database address: \n%s", err)
		return
	}

	// Create users and schemas and assign permissions to opencx
	if err = db.rootInitSchemas(); err != nil {
		err = fmt.Errorf("Root could not initialize schemas: \n%s", err)
		return
	}

	// open db handle
	openString := fmt.Sprintf("%s:%s@%s(%s)/", db.dbUsername, db.dbPassword, db.dbAddr.Network(), db.dbAddr.String())

	if db.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database: \n%s", err)
		return
	}

	// Get all the pairs
	if db.pairsArray, err = match.GenerateAssetPairs(coinList); err != nil {
		return
	}

	if err = db.setupSchemasAndTables(); err != nil {
		err = fmt.Errorf("Error setting up schemas and tables for setupclient: %s", err)
		return
	}

	return
}

func (db *DB) doesPairExist(newPair *match.Pair) (res bool) {
	for _, pair := range db.pairsArray {
		if bytes.Equal(pair.Serialize(), newPair.Serialize()) {
			res = true
		}
	}
	return
}

func (db *DB) setupSchemasAndTables() (err error) {

	if err = db.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	if err = db.SetupCustodyTables(db.balanceSchema, db.depositSchema, db.pendingDepositSchema); err != nil {
		err = fmt.Errorf("Error setting up custody tables: %s", err)
		return
	}

	if err = db.SetupExchangeTables(db.orderSchema); err != nil {
		err = fmt.Errorf("Error setting up exchange tables: %s", err)
		return
	}

	if err = db.SetupAuctionTables(db.auctionSchema, db.puzzleSchema, db.puzzleTable, db.auctionOrderSchema, db.auctionOrderTable); err != nil {
		err = fmt.Errorf("Error setting up auction tables: %s", err)
		return
	}

	return
}

// SetupPeerTables sets up tables required for the database to conform to Lit Peer Storage
func (db *DB) SetupPeerTables(peerSchema string, peerTable string) (err error) {

	// peer schema stuff
	// Initialize peer table
	// TODO: change this when peeridx is deprecated, if it ever is
	if err = db.InitializeSingleTable(db.peerSchema, db.peerTableName, "lnaddr VARBINARY(40), name TEXT, netaddr TEXT, peerIdx INT(32) UNSIGNED"); err != nil {
		err = fmt.Errorf("Could not initialize peer tables: \n%s", err)
		return
	}

	return
}

// SetupCustodyTables sets up the tables needed to track what funds a user has
func (db *DB) SetupCustodyTables(balanceSchema string, depositSchema string, pendingDepositSchema string) (err error) {

	// Initialize Balance tables
	// 66 bytes because we use big bytes and they use small bytes for varbinary
	if err = db.InitializeTables(balanceSchema, "pubkey VARBINARY(66), balance BIGINT(64)"); err != nil {
		err = fmt.Errorf("Could not initialize balance tables: \n%s", err)
		return
	}

	// Initialize Deposit tables
	if err = db.InitializeTables(depositSchema, "pubkey VARBINARY(66), address VARCHAR(34), CONSTRAINT unique_pubkeys UNIQUE (pubkey, address)"); err != nil {
		err = fmt.Errorf("Could not initialize deposit tables: \n%s", err)
		return
	}

	// Initialize pending_deposits table
	if err = db.InitializeNewTables(pendingDepositSchema, "pubkey VARBINARY(66), expectedConfirmHeight INT(32) UNSIGNED, depositHeight INT(32) UNSIGNED, amount BIGINT(64), txid TEXT"); err != nil {
		err = fmt.Errorf("Could not initialize pending deposit tables: \n%s", err)
		return
	}

	return
}

// SetupExchangeTables sets up the tables needed for an orderbook
func (db *DB) SetupExchangeTables(orderSchema string) (err error) {
	// Initialize order table
	// You can have a price up to 30 digits total, and 10 decimal places.
	if err = db.InitializePairTables(orderSchema, "pubkey VARBINARY(66), orderID TEXT, side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"); err != nil {
		err = fmt.Errorf("Could not initialize order tables: \n%s", err)
		return
	}
	return
}

// SetupAuctionTables sets up the tables needed to store auction orders and puzzles for specific auctions
func (db *DB) SetupAuctionTables(auctionSchema string, puzzleSchema string, puzzleTable string, auctionOrderSchema string, auctionOrderTable string) (err error) {

	// Initialize auction order schema, table
	// An auction order is identified by it's auction ID, pubkey, nonce, and other specific data.
	// You can have a price up to 30 digits total, and 10 decimal places.
	// Could be using blob instead of text but it really doesn't matter
	if err = db.InitializePairTables(auctionSchema, "pubkey VARBINARY(66), side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), auctionID VARBINARY(64), nonce VARBINARY(4), sig BLOB, hashedOrder VARBINARY(64), PRIMARY KEY (hashedOrder)"); err != nil {
		err = fmt.Errorf("Could not initialize auction order tables: \n%s", err)
		return
	}

	// This creates the single table where we'll keep all the puzzles
	if err = db.InitializeSingleTable(puzzleSchema, puzzleTable, "encodedOrder TEXT, auctionID VARBINARY(64), selected BOOLEAN"); err != nil {
		err = fmt.Errorf("Could not initialize puzzle table: \n%s", err)
		return
	}

	// This creates the single table where we'll keep the mapping of auction ID to auction number
	if err = db.InitializeSingleTable(auctionOrderSchema, auctionOrderTable, "auctionID VARBINARY(64), auctionNumber BIGINT(64), PRIMARY KEY (auctionNumber)"); err != nil {
		err = fmt.Errorf("Could not initialize auction order map table: %s", err)
		return
	}

	return
}

// InitializeSingleTable initializes a single table in a schema
func (db *DB) InitializeSingleTable(schemaName string, tableName string, schemaSpec string) (err error) {

	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}

	// now create the table
	tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, schemaSpec)
	if _, err = db.DBHandler.Exec(tableQuery); err != nil {
		err = fmt.Errorf("Could not create table %s: \n%s", tableName, err)
		return
	}
	return
}

// InitializeTables initializes all of the tables necessary for the exchange to run, using the coin list that is set up.
func (db *DB) InitializeTables(schemaName string, schemaSpec string) (err error) {

	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}
	for _, chain := range db.coinList {
		tableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", chain.Name, schemaSpec)
		if _, err = db.DBHandler.Exec(tableQuery); err != nil {
			err = fmt.Errorf("Could not create table %s: \n%s", chain.Name, err)
			return
		}
	}
	return
}

// InitializeNewTables initializes tables based on schema and clears them.
func (db *DB) InitializeNewTables(schemaName string, schemaSpec string) (err error) {
	// Use the schema
	if _, err = db.DBHandler.Exec("USE " + schemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: \n%s", schemaName, err)
		return
	}
	for _, chain := range db.coinList {
		dropTableQuery := fmt.Sprintf("DROP TABLE IF EXISTS %s;", chain.Name)
		if _, err = db.DBHandler.Exec(dropTableQuery); err != nil {
			err = fmt.Errorf("Could not drop table %s for initnewtables: \n%s", chain.Name, err)
			return
		}
		createTableQuery := fmt.Sprintf("CREATE TABLE %s (%s);", chain.Name, schemaSpec)
		if _, err = db.DBHandler.Exec(createTableQuery); err != nil {
			err = fmt.Errorf("Could not create table %s: \n%s", chain.Name, err)
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

// rootInitSchemas initializes the schemas, creates users, and grants permissions to those users
func (db *DB) rootInitSchemas() (err error) {

	// open db handle
	openString := fmt.Sprintf("%s:%s@%s(%s)/", db.dbUsername, db.dbPassword, db.dbAddr.Network(), db.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database: \n%s", err)
		return
	}

	// When the method is done, close the root connection
	defer rootHandler.Close()

	if err = rootHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: \n%s", err)
		return
	}

	schemasToCreate := []string{
		db.balanceSchema,
		db.depositSchema,
		db.pendingDepositSchema,
		db.orderSchema,
		db.peerSchema,
		db.puzzleSchema,
		db.auctionSchema,
		db.auctionOrderSchema,
	}

	for _, schema := range schemasToCreate {
		if err = rootCreateSchemaForUser(rootHandler, db.dbUsername, schema); err != nil {
			err = fmt.Errorf("Error calling rootCreateSchemaForUser helper: \n%s", err)
			return
		}

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
