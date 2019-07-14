package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"
	"testing"

	"github.com/mit-dci/lit/coinparam"
)

var (
	testStandardAuctionTime = uint64(100)
	// normal db user stuff
	testingUser = "testopencx"
	testingPass = "testpass"
	// root user stuff
	rootUser = "root"
	rootPass = ""
	// test string to put before test schemas
	testString = "testopencxdb_"
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

	var dbAddr net.Addr
	if dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(testConfig().DBHost, fmt.Sprintf("%d", testConfig().DBPort))); err != nil {
		t.Errorf("Error resolving conf derived address for killDatabaseFunc: %s", err)
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbAddr.Network(), dbAddr.String())

	// this is the root user!
	var dbHandle *sql.DB
	if dbHandle, err = sql.Open("mysql", openString); err != nil {
		t.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// make sure we close the connection at the end
	defer dbHandle.Close()

	for _, schema := range getSchemasFromConfig(testConfig()) {
		if schema != "" {
			if _, err = dbHandle.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", schema)); err != nil {
				t.Errorf("Error dropping db for testing: %s", err)
				return
			}
		}
	}

	return
}

// takes in a t so we can log things
func createUserAndDatabaseBench() (killThemBoth func(b *testing.B), err error) {

	var killUserFunc func() (err error)

	// create user first, then return killer
	if killUserFunc, err = createOpencxUser(); err != nil {
		err = fmt.Errorf("Error creating opencx user while creating both: %s", err)
		return
	}

	killThemBoth = func(b *testing.B) {
		// kill user first then database
		if err = killUserFunc(); err != nil {
			b.Errorf("Error killing user while killing both: %s", err)
			return
		}
		killDatabaseFuncBench(b)
		return
	}
	return
}

func killDatabaseFuncBench(b *testing.B) {
	var err error

	var dbAddr net.Addr
	if dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(testConfig().DBHost, fmt.Sprintf("%d", testConfig().DBPort))); err != nil {
		b.Errorf("Error resolving conf derived address for killDatabaseFunc: %s", err)
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbAddr.Network(), dbAddr.String())

	// this is the root user!
	var dbHandle *sql.DB
	if dbHandle, err = sql.Open("mysql", openString); err != nil {
		b.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// make sure we close the connection at the end
	defer dbHandle.Close()

	for _, schema := range getSchemasFromConfig(testConfig()) {
		if schema != "" {
			if _, err = dbHandle.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", schema)); err != nil {
				b.Errorf("Error dropping db for testing: %s", err)
				return
			}
		}
	}

	return
}
func getSchemasFromConfig(conf *dbsqlConfig) (schemas []string) {
	return []string{
		conf.PuzzleSchemaName,
		conf.AuctionOrderSchemaName,
		conf.AuctionSchemaName,
		conf.ReadOnlyBalanceSchemaName,
		conf.ReadOnlyAuctionSchemaName,
		conf.PendingDepositSchemaName,
		conf.ReadOnlyOrderSchemaName,
		conf.DepositSchemaName,
		conf.BalanceSchemaName,
		conf.OrderSchemaName,
		conf.PeerSchemaName,
	}
}

// createOpencxUser creates a user with the root user. If it can't do that then we need to be able to skip the test.
func createOpencxUser() (killUserFunc func() (err error), err error) {

	var dbAddr net.Addr
	if dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(testConfig().DBHost, fmt.Sprintf("%d", testConfig().DBPort))); err != nil {
		err = fmt.Errorf("Error resolving conf derived address for killDatabaseFunc: %s", err)
		return
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbAddr.Network(), dbAddr)

	// this is the root user!
	var dbHandle *sql.DB
	if dbHandle, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = dbHandle.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// make sure we close the connection at the end
	defer dbHandle.Close()

	if _, err = dbHandle.Exec(fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, CREATE, DROP, DELETE ON *.* TO '%s'@'%s' IDENTIFIED BY '%s';", testConfig().DBUsername, testConfig().DBHost, testConfig().DBPassword)); err != nil {
		err = fmt.Errorf("Error creating user for testing: %s", err)
		return
	}

	// we pass them the function to kill the user
	killUserFunc = func() (err error) {

		// create open string for db
		openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbAddr.Network(), dbAddr)

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
