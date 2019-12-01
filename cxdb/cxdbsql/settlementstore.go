package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"
	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/match"
)

// The SQLSettlementStore handles all client-viewable transactions relating to settlement.
// The difference between SettlementStore and SettlementEngine:
// SettlementStore is meant to be updated when the settlement engine returns, the settlement engine
// is the thing that actually handles validation.
// We do not care what this says if we want to place an order.
type SQLSettlementStore struct {
	DBHandler *sql.DB

	// db username
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// balance schema name
	balanceReadOnlySchema string

	// this coin
	coin *coinparam.Params
}

const (
	settlementStoreSchema = "pubkey VARBINARY(66), balance BIGINT(64), PRIMARY KEY (pubkey)"
)

// CreateSettlementStore creates a settlement store for a specific coin.
func CreateSettlementStore(coin *coinparam.Params) (store cxdb.SettlementStore, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// Set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateSQLSettlementStore: %s", err)
		return
	}

	// Set values
	ss := &SQLSettlementStore{
		dbUsername:            conf.DBUsername,
		dbPassword:            conf.DBPassword,
		balanceReadOnlySchema: conf.ReadOnlyBalanceSchemaName,
		dbAddr:                addr,
		coin:                  coin,
	}

	if err = ss.setupSettlementStoreTables(); err != nil {
		err = fmt.Errorf("Error setting up settlement store tables while creating store: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", ss.dbUsername, ss.dbPassword, ss.dbAddr.Network(), ss.dbAddr.String())
	if ss.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateSQLSettlementStore: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = ss.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// Now we actually set what we want
	store = ss
	return
}

// setupSettlementStoreTables sets up the tables needed for the auction orderbook.
// This assumes the schema name is set
func (ss *SQLSettlementStore) setupSettlementStoreTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", ss.dbUsername, ss.dbPassword, ss.dbAddr.Network(), ss.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup settlement store tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup settlement store tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while creating settlement store tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + ss.balanceReadOnlySchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup settlement store tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + ss.balanceReadOnlySchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", ss.balanceReadOnlySchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", ss.coin.Name, settlementStoreSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating settlement store table: %s", err)
		return
	}
	return
}

// UpdateBalances updates the balances from the settlement executions
func (ss *SQLSettlementStore) UpdateBalances(settlementResults []*match.SettlementResult) (err error) {
	// Now get asset from coin
	var assetForBal match.Asset
	if assetForBal, err = match.AssetFromCoinParam(ss.coin); err != nil {
		err = fmt.Errorf("Error getting asset from coin param: %s", err)
		return
	}

	// First create transaction
	var tx *sql.Tx
	if tx, err = ss.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while applying settlement exec: \n%s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while applying settlement exec: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use balance schema
	if _, err = tx.Exec("USE " + ss.balanceReadOnlySchema + ";"); err != nil {
		err = fmt.Errorf("Error using balance schema for GetBalance: %s", err)
		return
	}

	for _, setResult := range settlementResults {
		newBalQuery := fmt.Sprintf("INSERT INTO %s (balance, pubkey) VALUES (%d, '%x') ON DUPLICATE KEY UPDATE  balance='%[2]d';", assetForBal, setResult.NewBal, setResult.SuccessfulExec.Pubkey[:])
		if _, err = tx.Exec(newBalQuery); err != nil {
			err = fmt.Errorf("Error applying insert for GetBalance: %s", err)
			return
		}
	}
	return
}

// GetBalance gets the balance for a pubkey and an asset.
func (ss *SQLSettlementStore) GetBalance(pubkey *koblitz.PublicKey) (balance uint64, err error) {
	// Get asset from coin
	var assetForBal match.Asset
	if assetForBal, err = match.AssetFromCoinParam(ss.coin); err != nil {
		err = fmt.Errorf("Error getting asset from coin param: %s", err)
		return
	}

	// Then create transaction
	var tx *sql.Tx
	if tx, err = ss.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while getting balance: \n%s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting balance: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use balance schema
	if _, err = tx.Exec("USE " + ss.balanceReadOnlySchema + ";"); err != nil {
		err = fmt.Errorf("Error using balance schema for GetBalance: %s", err)
		return
	}

	var row *sql.Row
	curBalQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", assetForBal, pubkey.SerializeCompressed())
	// errs deferred until scan
	row = tx.QueryRow(curBalQuery)

	if err = row.Scan(&balance); err != nil {
		err = fmt.Errorf("Error scanning when getting balance: %s", err)
		return
	}

	return
}

// CreateSettlementStoreMap creates a map of coin to settlement engine, given a list of coins.
func CreateSettlementStoreMap(coins []*coinparam.Params) (setMap map[*coinparam.Params]cxdb.SettlementStore, err error) {

	setMap = make(map[*coinparam.Params]cxdb.SettlementStore)
	var curSetEng cxdb.SettlementStore
	for _, coin := range coins {
		if curSetEng, err = CreateSettlementStore(coin); err != nil {
			err = fmt.Errorf("Error creating single settlement store while creating settlement store map: %s", err)
			return
		}
		setMap[coin] = curSetEng
	}

	return
}
