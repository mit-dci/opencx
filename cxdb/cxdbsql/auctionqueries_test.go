package cxdbsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/match"
)

const (
	testStandardAuctionTime = 100
	// normal db user stuff
	testingUser = "testopencx"
	testingPass = "testpass"
	// root user stuff -- should be default
	rootUser = "root"
	rootPass = ""
	//
	testDBName = "testopencxdb"
)

var (
	testAuctionOrder = &match.AuctionOrder{
		Pubkey:     [...]byte{0x02, 0xe7, 0xb7, 0xcf, 0xcf, 0x42, 0x2f, 0xdb, 0x68, 0x2c, 0x85, 0x02, 0xbf, 0x2e, 0xef, 0x9e, 0x2d, 0x87, 0x67, 0xf6, 0x14, 0x67, 0x41, 0x53, 0x4f, 0x37, 0x94, 0xe1, 0x40, 0xcc, 0xf9, 0xde, 0xb3},
		Nonce:      [2]byte{0x00, 0x00},
		AuctionID:  [32]byte{0xde, 0xad, 0xbe, 0xef},
		AmountWant: 100000,
		AmountHave: 10000,
		Side:       "buy",
		TradingPair: match.Pair{
			AssetWant: match.Asset(6),
			AssetHave: match.Asset(8),
		},
		Signature: []byte{0x1b, 0xd6, 0x0f, 0xd3, 0xec, 0x5b, 0x73, 0xad, 0xa9, 0x8a, 0x92, 0x79, 0x82, 0x0f, 0x8e, 0xab, 0xf8, 0x8f, 0x47, 0x6e, 0xc3, 0x15, 0x33, 0x72, 0xd9, 0x90, 0x51, 0x41, 0xfd, 0x0a, 0xa1, 0xa2, 0x4a, 0x73, 0x75, 0x4c, 0xa5, 0x28, 0x4a, 0xc2, 0xed, 0x5a, 0xe9, 0x33, 0x22, 0xf4, 0x41, 0x1f, 0x9d, 0xd1, 0x78, 0xb9, 0x17, 0xd4, 0xe9, 0x72, 0x51, 0x7f, 0x5b, 0xd7, 0xe5, 0x12, 0xe7, 0x69, 0xb0},
	}
	testEncryptedOrder, _ = testAuctionOrder.TurnIntoEncryptedOrder(testStandardAuctionTime)
	testEncryptedBytes, _ = testEncryptedOrder.Serialize()
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
		DBName:     testDBName,

		// schemas (test schema names)
		BalanceSchemaName:        defaultBalanceSchema,
		DepositSchemaName:        defaultDepositSchema,
		PendingDepositSchemaName: defaultPendingDepositSchema,
		PuzzleSchemaName:         defaultPuzzleSchema,
		AuctionSchemaName:        defaultAuctionSchema,
		AuctionOrderSchemaName:   defaultAuctionOrderSchema,
		OrderSchemaName:          defaultOrderSchema,
		PeerSchemaName:           defaultPeerSchema,

		// tables
		PuzzleTableName:       defaultPuzzleTable,
		AuctionOrderTableName: defaultAuctionOrderTable,
		PeerTableName:         defaultPeerTable,
	}
	return
}

func constCoinParams() (params []*coinparam.Params) {
	params = []*coinparam.Params{
		&coinparam.TestNet3Params,
		&coinparam.VertcoinTestNetParams,
		&coinparam.VertcoinRegTestParams,
		&coinparam.RegressionNetParams,
		&coinparam.LiteCoinTestNet4Params,
	}
	return
}

// takes in a t so we can log things
func createUserAndDatabase() (killThemBoth func(t *testing.T), err error) {

	var killUserFunc func() (err error)
	var killDatabaseFunc func() (err error)

	// create database first then user
	if killDatabaseFunc, err = createTestDatabase(); err != nil {
		err = fmt.Errorf("Error creating test database while creating both: %s", err)
		return
	}

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
		if err = killDatabaseFunc(); err != nil {
			t.Errorf("Error killing database while killing both: %s", err)
			return
		}
		return
	}
	return
}

func createTestDatabase() (killDatabaseFunc func() (err error), err error) {

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

	if _, err = dbHandle.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbConn.dbName)); err != nil {
		err = fmt.Errorf("Error creating user for testing: %s", err)
		return
	}

	killDatabaseFunc = func() (err error) {
		var dbConn *DB

		// initialize the db
		dbConn = new(DB)

		// We created this testConfig, now we set the options from the testConfig
		if err = dbConn.setOptionsFromConfig(testConfig()); err != nil {
			err = fmt.Errorf("Error setting options from testConfig for setupclient: %s", err)
			return
		}

		// create open string for db
		openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr.String())

		// this is the root user!
		var dbHandle *sql.DB
		if dbHandle, err = sql.Open("mysql", openString); err != nil {
			err = fmt.Errorf("Error opening db to create testing user: %s", err)
			return
		}

		// make sure we close the connection at the end
		defer dbHandle.Close()

		if _, err = dbHandle.Exec(fmt.Sprintf("DROP DATABASE %s;", dbConn.dbName)); err != nil {
			err = fmt.Errorf("Error dropping db for testing: %s", err)
			return
		}

		return
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
	openString := fmt.Sprintf("%s:%s@%s(%s)/%s", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr.String(), dbConn.dbName)

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
		openString := fmt.Sprintf("%s:%s@%s(%s)/%s", rootUser, rootPass, dbConn.dbAddr.Network(), dbConn.dbAddr.String(), testConfig().DBName)

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
	openString := fmt.Sprintf("%s:%s@%s(%s)/%s", db.dbUsername, db.dbPassword, db.dbAddr.Network(), db.dbAddr.String(), db.dbName)

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

// TestPlaceAuctionGoodParams should succeed with the correct coin params.
func TestPlaceAuctionGoodParams(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Skipf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("Error placing auction puzzle, should not error: %s", err)
		return
	}

	return
}

// TestPlaceAuctionPuzzleSuccess should succeed even with bad coin params.
func TestPlaceAuctionBadParams(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("There was no error placing auction puzzle, should not error even w bad params: %s", err)
		return
	}

	return
}

// TestPlaceAuctionOrderbookChanges should succeed with the correct coin params.
func TestPlaceAuctionOrderbookChanges(t *testing.T) {
	var err error

	// first create the user for the db
	var killThemBoth func(t *testing.T)
	if killThemBoth, err = createUserAndDatabase(); err != nil {
		t.Skipf("Could not create user for test (error), so skipping: %s", err)
		return
	}

	defer killThemBoth(t)

	var dbConn *DB
	if dbConn, err = startupDB(); err != nil {
		t.Errorf("Error starting db for place auction test: %s", err)
		return
	}

	if err = dbConn.PlaceAuctionPuzzle(testEncryptedOrder); err != nil {
		t.Errorf("Error placing auction puzzle, should not error: %s", err)
		return
	}

	// Starting from an empty book, we should see this order added.
	var returnedOrders []*match.EncryptedAuctionOrder
	if returnedOrders, err = dbConn.ViewAuctionPuzzleBook(testEncryptedOrder.IntendedAuction); err != nil {
		t.Errorf("Error vewing auction puzzle book, should not error: %s", err)
		return
	}

	if len(returnedOrders) != 1 {
		t.Errorf("Length of returned orders is %d, should be 1", len(returnedOrders))
		return
	}

	var retBytes []byte
	if retBytes, err = returnedOrders[0].Serialize(); err != nil {
		t.Errorf("Error serializing first returned order, should not error: %s", err)
		return
	}

	if bytes.Compare(retBytes, testEncryptedBytes) != 0 {
		t.Errorf("The serialized returned order from orderbook was not the same as the input, should be equal")
		return
	}

	return
}
