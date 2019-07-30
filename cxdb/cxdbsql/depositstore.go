package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
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
	depositAddrStoreSchema    = "pubkey VARBINARY(66), address VARCHAR(34), CONSTRAINT unique_pubkeys UNIQUE (pubkey, address)"
	pendingDepositStoreSchema = "pubkey VARBINARY(66), expectedConfirmHeight INT(32) UNSIGNED, depositHeight INT(32) UNSIGNED, amount BIGINT(64), txid TEXT"
)

func CreateDepositStoreStructWithConf(coin *coinparam.Params, conf *dbsqlConfig) (ds *SQLDepositStore, err error) {

	// set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateDepositStore: %s", err)
		return
	}

	// Set values for limit engine
	ds = &SQLDepositStore{
		dbUsername:               conf.DBUsername,
		dbPassword:               conf.DBPassword,
		depositAddrSchemaName:    conf.DepositSchemaName,
		pendingDepositSchemaName: conf.PendingDepositSchemaName,

		dbAddr: addr,
		coin:   coin,
	}

	if err = ds.setupDepositTables(); err != nil {
		err = fmt.Errorf("Error setting up deposit tables for CreateDepositStore: %s", err)
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
		err = fmt.Errorf("Could not ping the database, is it running? Did you set the correct username and password in sqldb.conf: %s", err)
		return
	}

	return
}

func CreateDepositStore(coin *coinparam.Params) (store cxdb.DepositStore, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	if store, err = CreateDepositStoreStructWithConf(coin, conf); err != nil {
		err = fmt.Errorf("Error creating deposit store struct for CreateDepositStore: %s", err)
		return
	}
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

// DestroyHandler closes the DB handler that we created, and makes it nil
func (ds *SQLDepositStore) DestroyHandler() (err error) {
	if ds.DBHandler == nil {
		err = fmt.Errorf("Error, cannot destroy nil handler, please create new deposit store")
		return
	}
	if err = ds.DBHandler.Close(); err != nil {
		err = fmt.Errorf("Error closing engine handler for DestroyHandler: %s", err)
		return
	}
	ds.DBHandler = nil
	return
}

// UpdateDeposits updates the deposits when a block comes in, and returns execs for deposits that are
// now confirmed
func (ds *SQLDepositStore) UpdateDeposits(deposits []match.Deposit, blockheight uint64) (depositExecs []*match.SettlementExecution, err error) {

	// first get debit asset
	var depositAsset match.Asset
	if depositAsset, err = match.AssetFromCoinParam(ds.coin); err != nil {
		err = fmt.Errorf("Error getting asset from coin param for UpdateDeposits: %s", err)
		return
	}

	// ACID
	var tx *sql.Tx
	if tx, err = ds.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for UpdateDeposits: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for UpdateDeposits: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the pending deposit schema
	if _, err = tx.Exec("USE " + ds.pendingDepositSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for UpdateDeposits: %s", err)
		return
	}

	// First we insert these deposits
	for _, deposit := range deposits {
		txid := []byte(deposit.Txid)
		expectedConfirm := deposit.BlockHeightReceived + deposit.Confirmations
		insertDepQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d, %d, %d, '%x');", ds.coin.Name, deposit.Pubkey.SerializeCompressed(), expectedConfirm, deposit.BlockHeightReceived, deposit.Amount, txid)
		if _, err = tx.Exec(insertDepQuery); err != nil {
			err = fmt.Errorf("Error inserting deposit for UpdateDeposits: %s", err)
			return
		}
	}

	// There is an issue with the way we do this, if a reorg were to happen, users would be credited
	// twice because there would be two blocks worth of deposits for some expected confirm heights.
	// This is why we use non custodial / lightning anyways, if my channel fund was confirmed 15 days ago
	// I really don't think it will be reorged. So for all intents and purposes this method of taking
	// money from people should be considered legacy.

	// Now we select the ones where expectedConfirm EQUALS the current height.
	var rows *sql.Rows
	selectConfirmedQuery := fmt.Sprintf("SELECT pubkey, amount FROM %s WHERE expectedConfirmHeight=%d;", ds.coin.Name, blockheight)
	if rows, err = tx.Query(selectConfirmedQuery); err != nil {
		err = fmt.Errorf("Error running select confirmed query for UpdateDeposits: %s", err)
		return
	}

	var currSettlement *match.SettlementExecution
	var pubkeyBytes []byte
	for rows.Next() {
		// A confirmed deposit is a debit for the deposit store's asset
		currSettlement = &match.SettlementExecution{
			Asset: depositAsset,
			Type:  match.Debit,
		}
		if err = rows.Scan(&pubkeyBytes, &currSettlement.Amount); err != nil {
			err = fmt.Errorf("Error scanning for confirmed deposit: %s", err)
			return
		}

		// because we really only know that sql will give us a hex string, not actual bytes
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			err = fmt.Errorf("Error decoding pubkey bytes string for UpdateDeposits: %s", err)
			return
		}
		copy(currSettlement.Pubkey[:], pubkeyBytes)
		// Now that the settlement is filled in, let's add it
		depositExecs = append(depositExecs, currSettlement)
	}
	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for UpdateDeposits: %s", err)
		return
	}

	return
}

// GetDepositAddressMap gets a map of the deposit addresses we own to pubkeys
func (ds *SQLDepositStore) GetDepositAddressMap() (depAddrMap map[string]*koblitz.PublicKey, err error) {
	depAddrMap = make(map[string]*koblitz.PublicKey)
	// ACID
	var tx *sql.Tx
	if tx, err = ds.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for GetDepositAddressMap: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for GetDepositAddressMap: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the deposit address schema
	if _, err = tx.Exec("USE " + ds.depositAddrSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for GetDepositAddressMap: %s", err)
		return
	}

	var rows *sql.Rows
	selectAddrQuery := fmt.Sprintf("SELECT pubkey, address FROM %s;", ds.coin.Name)
	// errors deferred to scan
	if rows, err = tx.Query(selectAddrQuery); err != nil {
		err = fmt.Errorf("Error querying for pubkey address map for GetDepositAddressMap: %s", err)
		return
	}

	var currAddr string
	var currPubkeyBytes []byte
	for rows.Next() {
		if err = rows.Scan(&currPubkeyBytes, &currAddr); err != nil {
			err = fmt.Errorf("Error scanning for address for GetDepositAddressMap: %s", err)
			return
		}

		// frustrating sql marshalling
		if currPubkeyBytes, err = hex.DecodeString(string(currPubkeyBytes)); err != nil {
			err = fmt.Errorf("Error decoding pubkey for GetDepositAddressMap: %s", err)
			return
		}

		if depAddrMap[currAddr], err = koblitz.ParsePubKey(currPubkeyBytes, koblitz.S256()); err != nil {
			err = fmt.Errorf("Error parsing pub key from bytes for GetDepositAddressMap: %s", err)
			return
		}
	}
	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for GetDepositAddressMap: %s", err)
		return
	}
	return
}

// GetDepositAddress gets the deposit address for a pubkey and an asset.
func (ds *SQLDepositStore) GetDepositAddress(pubkey *koblitz.PublicKey) (addr string, err error) {
	// ACID
	var tx *sql.Tx
	if tx, err = ds.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for GetDepositAddress: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for GetDepositAddress: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the deposit address schema
	if _, err = tx.Exec("USE " + ds.depositAddrSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for GetDepositAddress: %s", err)
		return
	}

	var row *sql.Row
	selectAddrQuery := fmt.Sprintf("SELECT address FROM %s WHERE pubkey='%s';", ds.coin.Name, pubkey.SerializeCompressed())
	// errors deferred to scan
	row = tx.QueryRow(selectAddrQuery)

	if err = row.Scan(&addr); err != nil {
		err = fmt.Errorf("Error scanning for address for GetDepositAddress: %s", err)
		return
	}

	return
}

// RegisterUser takes in a pubkey, and an address for the pubkey, and puts the deposit address as the
// value for the user's pubkey key
func (ds *SQLDepositStore) RegisterUser(pubkey *koblitz.PublicKey, address string) (err error) {
	// ACID
	var tx *sql.Tx
	if tx, err = ds.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for RegisterUser: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for RegisterUser: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the deposit address schema
	if _, err = tx.Exec("USE " + ds.depositAddrSchemaName + ";"); err != nil {
		err = fmt.Errorf("Error using puzzle schema for RegisterUser: %s", err)
		return
	}

	insertUserQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s');", ds.coin.Name, pubkey.SerializeCompressed(), address)
	if _, err = tx.Exec(insertUserQuery); err != nil {
		err = fmt.Errorf("Error adding user and address for RegisterUser: %s", err)
		return
	}
	return
}

// CreateDepositStoreMap creates a map of pair to deposit store, given a list of pairs.
func CreateDepositStoreMap(coinList []*coinparam.Params) (depositMap map[*coinparam.Params]cxdb.DepositStore, err error) {

	depositMap = make(map[*coinparam.Params]cxdb.DepositStore)
	var curDepositMap cxdb.DepositStore
	for _, coin := range coinList {
		if curDepositMap, err = CreateDepositStore(coin); err != nil {
			err = fmt.Errorf("Error creating single deposit store while creating deposit store map: %s", err)
			return
		}
		depositMap[coin] = curDepositMap
	}

	return
}
