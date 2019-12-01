package cxdbsql

import (
	"database/sql"
	"fmt"
	"net"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Rjected/lit/coinparam"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

type SQLSettlementEngine struct {
	DBHandler *sql.DB

	// db username
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// balance schema name
	balanceSchema string

	// this coin
	coin *coinparam.Params
}

const (
	settlementEngineSchema = "pubkey VARBINARY(66), balance BIGINT(64), PRIMARY KEY (pubkey)"
)

// CreateSettlementEngine creates a settlement engine for a specific coin
func CreateSettlementEngine(coin *coinparam.Params) (engine match.SettlementEngine, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// Set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateSQLSettlementEngine: %s", err)
		return
	}

	// Set values
	se := &SQLSettlementEngine{
		dbUsername:    conf.DBUsername,
		dbPassword:    conf.DBPassword,
		balanceSchema: conf.BalanceSchemaName,
		dbAddr:        addr,
		coin:          coin,
	}

	if err = se.setupSettlementTables(); err != nil {
		err = fmt.Errorf("Error setting up settlement engine tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", se.dbUsername, se.dbPassword, se.dbAddr.Network(), se.dbAddr.String())
	if se.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateSQLSettlementEngine: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = se.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// Now we actually set what we want
	engine = se
	return
}

// ApplySettlementExecution applies the settlementExecution, this assumes that the settlement execution is
// valid
func (se *SQLSettlementEngine) ApplySettlementExecution(setExec *match.SettlementExecution) (setRes *match.SettlementResult, err error) {

	// First create transaction
	var tx *sql.Tx
	if tx, err = se.DBHandler.Begin(); err != nil {
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
	if _, err = tx.Exec("USE " + se.balanceSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	curBalQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", se.coin.Name, setExec.Pubkey)
	if rows, err = tx.Query(curBalQuery); err != nil {
		err = fmt.Errorf("Error querying for balance while applying settlement exec: %s", err)
		return
	}

	var curBal uint64
	if rows.Next() {
		if err = rows.Scan(&curBal); err != nil {
			err = fmt.Errorf("Error scanning when applying settlement exec: %s", err)
			return
		}
	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for ApplySettlementExecution: %s", err)
		return
	}

	var newBal uint64
	if setExec.Type == match.Debit {
		newBal = curBal + setExec.Amount
	} else if setExec.Type == match.Credit {
		newBal = curBal - setExec.Amount
	}
	newBalQuery := fmt.Sprintf("INSERT INTO %s (balance, pubkey) VALUES (%d, '%x') ON DUPLICATE KEY UPDATE  balance='%[2]d';", se.coin.Name, newBal, setExec.Pubkey)
	if _, err = tx.Exec(newBalQuery); err != nil {
		err = fmt.Errorf("Error applying settlement exec new bal query: %s", err)
		return
	}

	// Finally set return value
	setRes = &match.SettlementResult{
		NewBal:         newBal,
		SuccessfulExec: setExec,
	}

	return
}

// CheckValid returns true if the settlement execution would be valid
func (se *SQLSettlementEngine) CheckValid(setExec *match.SettlementExecution) (valid bool, err error) {
	if setExec.Type == match.Debit {
		// No settlement will be an invalid debit
		valid = true
		return
	}
	// since we just returned, the setExec type == match.Credit

	// First create transaction
	var tx *sql.Tx
	if tx, err = se.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while checking settlement exec: \n%s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while checking settlement exec: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use balance schema
	if _, err = tx.Exec("USE " + se.balanceSchema + ";"); err != nil {
		err = fmt.Errorf("Error using balance schema for CheckValid: %s", err)
		return
	}

	var row *sql.Row
	curBalQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", setExec.Asset, setExec.Pubkey)
	// error deferred to scan
	row = tx.QueryRow(curBalQuery)

	var curBal uint64
	if err = row.Scan(&curBal); err != nil {
		err = fmt.Errorf("Error scanning when checking settlement exec: %s", err)
		return
	}

	logging.Infof("User with %d %s trying to complete action costing %d %[2]s.", curBal, se.coin.Name, setExec.Amount)
	valid = setExec.Amount < curBal
	return
}

// setupSettlementTables sets up the tables needed for the auction orderbook.
// This assumes the schema name is set
func (se *SQLSettlementEngine) setupSettlementTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", se.dbUsername, se.dbPassword, se.dbAddr.Network(), se.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup settlement tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup settlement tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while setting up settlement tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + se.balanceSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup settlement tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + se.balanceSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", se.balanceSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", se.coin.Name, settlementEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating settlement table: %s", err)
		return
	}
	return
}

// CreateSettlementEngineMap creates a map of coin to settlement engine, given a list of coins.
func CreateSettlementEngineMap(coins []*coinparam.Params) (setMap map[*coinparam.Params]match.SettlementEngine, err error) {

	setMap = make(map[*coinparam.Params]match.SettlementEngine)
	var curSetEng match.SettlementEngine
	for _, coin := range coins {
		if curSetEng, err = CreateSettlementEngine(coin); err != nil {
			err = fmt.Errorf("Error creating single settlement engine while creating settlement engine map: %s", err)
			return
		}
		setMap[coin] = curSetEng
	}

	return
}
