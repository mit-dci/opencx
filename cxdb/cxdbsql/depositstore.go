package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

type SQLDepositStore struct {
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// deposit addr schema name
	depositAddrSchemaName string

	// pending deposit schema name
	pendingDepositSchemaName string

	// this coin
	coin *coinparam.Params
}

// The schema for the deposit store
const (
	depositAddrStoreSchema    = "pubkey VARBINARY(66), address VARCHAR(34), CONSTRAINT unique_pubkeys UNIQUE (pubkeys, address)"
	pendingDepositStoreSchema = "pubkey VARBINARY(66), expectedConfirmHeight INT(32) UNSIGNED, depositHeight INT(32) UNSIGNED, amount BIGINT(64), txid TEXT"
)

func CreateDepositStore(coin *coinparam.Params) (store cxdb.DepositStore, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateDepositStore: %s", err)
		return
	}

	// Set values for limit engine
	ds := &SQLDepositStore{
		dbUsername:               conf.DBUsername,
		dbPassword:               conf.DBPassword,
		depositAddrSchemaName:    conf.DepositSchemaName,
		pendingDepositSchemaName: conf.PendingDepositSchemaName,

		dbAddr: addr,
		coin:   coin,
	}

	if err = ds.setupDepositTables(); err != nil {
		err = fmt.Errorf("Error setting up deposit tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", ds.dbUsername, ds.dbPassword, ds.dbAddr.Network(), ds.dbAddr.String())
	if ds.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateDepositStore: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = ds.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// Now we actually set what we want
	store = ds
	return
}

// setupDepositTables sets up the tables needed to keep track of pending deposits and deposit addresses.
// This assumes everything else is set
func (ds *SQLDepositStore) setupDepositTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", ds.dbUsername, ds.dbPassword, ds.dbAddr.Network(), ds.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup deposit tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup deposit tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while matching setup deposit tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the first schema (keeping track of deposit addresses)
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + ds.depositAddrSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup deposit addr tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + ds.depositAddrSchemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", ds.depositAddrSchemaName, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", ds.coin.Name, depositAddrStoreSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating deposit addr table: %s", err)
		return
	}

	// Now create the other schema (keeping track of pending deposits)
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + ds.pendingDepositSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup pending deposit tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + ds.pendingDepositSchemaName + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", ds.pendingDepositSchemaName, err)
		return
	}

	createTableQuery = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", ds.coin.Name, pendingDepositStoreSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating deposit addr table: %s", err)
		return
	}
	return
}

// UpdateDeposits updates the deposits when a block comes in
func (ds *SQLDepositStore) UpdateDeposits(deposits []match.Deposit, blockheight uint64) (depositExecs []*match.SettlementExecution, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
func (ds *SQLDepositStore) GetDepositAddressMap() (depAddrMap map[string]*koblitz.PublicKey, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetDepositAddress gets the deposit address for a pubkey and an asset.
func (ds *SQLDepositStore) GetDepositAddress(pubkey *koblitz.PublicKey) (addr string, err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// RegisterUser takes in a pubkey, and an address for the pubkey, and puts the deposit address as the
// value for the user's pubkey key
func (ds *SQLDepositStore) RegisterUser(pubkey *koblitz.PublicKey, address string) (err error) {
	// TODO
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// CreateDepositStoreMap creates a map of pair to deposit store, given a list of pairs.
func CreateDepositStoreMap(coinList []*coinparam.Params) (depositMap map[*coinparam.Params]cxdb.DepositStore, err error) {

	depositMap = make(map[*coinparam.Params]cxdb.DepositStore)
	var curDepositMap cxdb.DepositStore
	for _, coin := range coinList {
		if curDepositMap, err = CreateDepositStore(coin); err != nil {
			err = fmt.Errorf("Error creating single limit engine while creating limit engine map: %s", err)
			return
		}
		depositMap[coin] = curDepositMap
	}

	return
}
