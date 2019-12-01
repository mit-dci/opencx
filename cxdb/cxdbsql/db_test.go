package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"

	"github.com/Rjected/lit/coinparam"
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

type testerContainer struct {
	rootHandler *sql.DB
}

// CreateTesterContainer creates a struct that contains a SQL *DB, which should be an active SQL database connection that is meant for dropping databases created by the auction engine, creating a test user, and maintains a root connection.
func CreateTesterContainer() (tc *testerContainer, err error) {
	tc = new(testerContainer)
	var dbAddr net.Addr
	if dbAddr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(testConfig().DBHost, fmt.Sprintf("%d", testConfig().DBPort))); err != nil {
		err = fmt.Errorf("Error resolving conf derived address for killDatabaseFunc: %s", err)
		return
	}

	// create open string for db
	openString := fmt.Sprintf("%s:%s@%s(%s)/", rootUser, rootPass, dbAddr.Network(), dbAddr)

	// this is the root user!
	if tc.rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening db to create testing user: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = tc.rootHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	if _, err = tc.rootHandler.Exec(fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, CREATE, DROP, DELETE ON *.* TO '%s'@'%s' IDENTIFIED BY '%s';", testConfig().DBUsername, testConfig().DBHost, testConfig().DBPassword)); err != nil {
		err = fmt.Errorf("Error creating user for testing: %s", err)
		return
	}
	return
}

// DropDBs drops the databases that would have been created by the test config if they exist
func (tc *testerContainer) DropDBs() (err error) {
	if tc.rootHandler == nil {
		err = fmt.Errorf("Error, cannot use nil handler, construct container correctly")
		return
	}
	for _, schema := range getSchemasFromConfig(testConfig()) {
		if schema != "" {
			if _, err = tc.rootHandler.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s;", schema)); err != nil {
				err = fmt.Errorf("Error dropping db for testing: %s", err)
				return
			}
		}
	}
	return
}

// Kill runs KillUser, then DropDBs, then closes the handler
func (tc *testerContainer) Kill() (err error) {
	if tc.rootHandler == nil {
		err = fmt.Errorf("Error, cannot kill nil handler, construct container correctly")
		return
	}
	if err = tc.KillUser(); err != nil {
		err = fmt.Errorf("Error killing user for Kill: %s", err)
		return
	}
	if err = tc.DropDBs(); err != nil {
		err = fmt.Errorf("Error dropping DBs for Kill: %s", err)
		return
	}
	if err = tc.rootHandler.Close(); err != nil {
		err = fmt.Errorf("Error closing tc handler for Kill: %s", err)
		return
	}
	return
}

// KillUser drops the user that should have been created when the testerContainer was created.
func (tc *testerContainer) KillUser() (err error) {
	if tc.rootHandler == nil {
		err = fmt.Errorf("Error killing user, cannot have nil handler, construct container correctly")
		return
	}
	if _, err = tc.rootHandler.Exec(fmt.Sprintf("DROP USER '%s'@'%s';", testingUser, testConfig().DBHost)); err != nil {
		err = fmt.Errorf("Error dropping user for testing: %s", err)
		return
	}

	return
}

// CloseHandler closes the handler for the test container, checking first if it's nil, and if it is, returning an error.
// CloseHandler also will pass along handler errors when closing.
func (tc *testerContainer) CloseHandler() (err error) {
	if tc.rootHandler == nil {
		err = fmt.Errorf("Error, cannot close nil handler, construct container correctly")
		return
	}
	if err = tc.rootHandler.Close(); err != nil {
		err = fmt.Errorf("Error closing tc handler for CloseHandler: %s", err)
		return
	}
	return
}

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
