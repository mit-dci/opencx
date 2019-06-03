package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

func testConfig() (conf *dbsqlConfig) {
	conf = &dbsqlConfig{
		// home dir (for test stuff)
		DBHomeDir: defaultDBHomeDirName + "test/",

		// user / pass / net stuff
		DBUsername: testingUser,
		DBPassword: testingPass,
		DBHost:     defaultDBHost,
		DBPort:     defaultDBPort,

		// schemas (test schema names)
		BalanceSchemaName:        testString + defaultBalanceSchema,
		DepositSchemaName:        testString + defaultDepositSchema,
		PendingDepositSchemaName: testString + defaultPendingDepositSchema,
		PuzzleSchemaName:         testString + defaultPuzzleSchema,
		AuctionSchemaName:        testString + defaultAuctionSchema,
		AuctionOrderSchemaName:   testString + defaultAuctionOrderSchema,
		OrderSchemaName:          testString + defaultOrderSchema,
		PeerSchemaName:           testString + defaultPeerSchema,

		// tables
		PuzzleTableName:       testString + defaultPuzzleTable,
		AuctionOrderTableName: testString + defaultAuctionOrderTable,
		PeerTableName:         testString + defaultPeerTable,
	}
	return
}

func getSchemasFromDB(db *DB) (testSchemaList []string, err error) {
	if db == nil {
		err = fmt.Errorf("Cannot get schemas from null db pointer: %s", err)
		return
	}
	testSchemaList = []string{
		db.balanceSchema,
		db.depositSchema,
		db.pendingDepositSchema,
		db.orderSchema,
		db.peerSchema,
		db.puzzleSchema,
		db.auctionSchema,
		db.auctionOrderSchema,
	}
	return
}

func constCoinParams() (params []*coinparam.Params) {
	params = []*coinparam.Params{
		&coinparam.TestNet3Params,
		&coinparam.VertcoinTestNetParams,
		&coinparam.VertcoinRegTestParams,
		&coinparam.RegressionNetParams,
		&coinparam.LiteRegNetParams,
		&coinparam.LiteCoinTestNet4Params,
	}
	return
}

// takes in a t so we can log things
func createUserAndDatabase() (killThemBoth func(t *testing.T), err error) {

	var killUserFunc func() (err error)

	// create user first, then return killer
	if killUserFunc, err = createOpencxUser(); err != nil {
		err = fmt.Errorf("Error creating opencx user while creating both: %s", err)
		return
	}

	killThemBoth = func(t *testing.T) {
		// kill user first then database
		if err = killUserFunc(); err != nil {
			t.Errorf("Error killing user while killing both: %s", err)
			return
		}
		killDatabaseFunc(t)
		return
	}
	return
}

func killDatabaseFunc(t *testing.T) {
	var err error
	var dbConn *DB

	// initialize the db
	dbConn = new(DB)

	// We created this testConfig, now we set the options from the testConfig
	if err = dbConn.setOptionsFromConfig(testConfig()); err != nil {
		t.Errorf("Error setting options from testConfig for setupclient: %s", err)
		return
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr.String())

	// this is the root user!
	var dbHandle *sql.DB
	if dbHandle, err = sql.Open("mysql", openString); err != nil {
		t.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// make sure we close the connection at the end
	defer dbHandle.Close()

	var schemaList []string
	if schemaList, err = getSchemasFromDB(dbConn); err != nil {
		t.Errorf("Error getting schemas from DB: %s", err)
		return
	}
	for _, schema := range schemaList {
		if _, err = dbHandle.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", schema)); err != nil {
			t.Errorf("Error dropping db for testing: %s", err)
			return
		}
	}

	return
}

// createOpencxUser creates a user with the root user. If it can't do that then we need to be able to skip the test.
func createOpencxUser() (killUserFunc func() (err error), err error) {

	var dbConn *DB

	// initialize the db
	dbConn = new(DB)

	// We created this testConfig, now we set the options from the testConfig
	if err = dbConn.setOptionsFromConfig(testConfig()); err != nil {
		err = fmt.Errorf("Error setting options from testConfig for setupclient: %s", err)
		return
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr)

	// this is the root user!
	var dbHandle *sql.DB
	if dbHandle, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// make sure we close the connection at the end
	defer dbHandle.Close()

	if _, err = dbHandle.Exec(fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, CREATE, DROP, DELETE ON *.* TO '%s'@'%s' IDENTIFIED BY '%s';", dbConn.dbUsername, testConfig().DBHost, dbConn.dbPassword)); err != nil {
		err = fmt.Errorf("Error creating user for testing: %s", err)
		return
	}

	// we pass them the function to kill the user
	killUserFunc = func() (err error) {
		var dbConn *DB

		// initialize the db
		dbConn = new(DB)

		// We created this testConfig, now we set the options from the testConfig
		if err = dbConn.setOptionsFromConfig(testConfig()); err != nil {
			err = fmt.Errorf("Error setting options from testConfig for setupclient: %s", err)
			return
		}

		// create open string for db
		openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr)

		// this is the root user!
		var dbHandle *sql.DB
		if dbHandle, err = sql.Open("mysql", openString); err != nil {
			err = fmt.Errorf("Error opening db to create testing user: %s", err)
			return
		}

		// make sure we close the connection at the end
		defer dbHandle.Close()

		if _, err = dbHandle.Exec(fmt.Sprintf("DROP USER '%s'@'%s';", testingUser, testConfig().DBHost)); err != nil {
			err = fmt.Errorf("Error dropping user for testing: %s", err)
			return
		}

		return
	}

	return
}

// startupDB starts up a test db client
func startupDB() (db *DB, err error) {

	// initialize the db
	db = new(DB)

	// We created this testConfig, now we set the options from the testConfig
	if err = db.setOptionsFromConfig(testConfig()); err != nil {
		err = fmt.Errorf("Error setting options from testConfig for setupclient: %s", err)
		return
	}

	db.gPriceMap = make(map[string]float64)
	db.priceMapMtx = new(sync.Mutex)
	db.coinList = constCoinParams()

	// Resolve address for host and port, then set that as the network address
	if db.dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(testConfig().DBHost, fmt.Sprintf("%d", testConfig().DBPort))); err != nil {
		err = fmt.Errorf("Error resolving database address: \n%s", err)
		return
	}

	// Create users and schemas and assign permissions to opencx
	if err = db.rootInitSchemas(); err != nil {
		err = fmt.Errorf("Root could not initialize schemas: \n%s", err)
		return
	}

	// open db handle
	openString := fmt.Sprintf("%s:%s@%s(%s)/", db.dbUsername, db.dbPassword, db.dbAddr.Network(), db.dbAddr)

	if db.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database: \n%s", err)
		return
	}

	// Get all the pairs
	if db.pairsArray, err = match.GenerateAssetPairs(constCoinParams()); err != nil {
		return
	}

	if err = db.setupSchemasAndTables(); err != nil {
		err = fmt.Errorf("Error setting up schemas and tables for setupclient: %s", err)
		return
	}

	return
}
