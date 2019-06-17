package cxdbsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/opencx/match"
)

// ApplySettlementExecution applies the settlementExecution, this assumes that the settlement execution is
// valid
func (db *DB) ApplySettlementExecution(setExec *match.SettlementExecution) (err error) {

	// First create transaction
	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
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
	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	curBalQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", setExec.Asset, setExec.Pubkey)
	if rows, err = tx.Query(curBalQuery); err != nil {
		err = fmt.Errorf("Error querying for balance while applying settlement exec: %s")
		return
	}

	var curBal uint64
	if rows.Next() {
		if err = rows.Scan(&curBal); err != nil {
			err = fmt.Errorf("Error scanning when applying settlement exec: %s", err)
			return
		}
	}

	var newBal uint64
	if setExec.Type == match.Debit {
		newBal = curBal + setExec.Amount
	} else if setExec.Type == Credit {
		newBal = curBal - setExec.Amount
	}
	newBalQuery := fmt.Sprintf("REPLACE INTO %s balance VALUES (%d);", setExec.Asset, newBal)
	if _, err = tx.Exec(newBalQuery); err != nil {
		err = fmt.Errorf("Error applying settlement exec new bal query: %s", err)
		return
	}

	return
}

// CheckValid returns true if the settlement execution would be valid
func (db *DB) CheckValid(setExec *match.SettlementExecution) (valid bool, err error) {
	if setExec.Type == match.Debit {
		// No settlement will be an invalid debit
		valid = true
		return
	}
	// since we just returned, the setExec type == match.Credit

	// First create transaction
	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
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
	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	curBalQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", setExec.Asset, setExec.Pubkey)
	if rows, err = tx.Query(curBalQuery); err != nil {
		err = fmt.Errorf("Error querying for balance while applying settlement exec: %s")
		return
	}

	var curBal uint64
	if rows.Next() {
		if err = rows.Scan(&curBal); err != nil {
			err = fmt.Errorf("Error scanning when applying settlement exec: %s", err)
			return
		}
	}

	valid = setExec.Amount > curBal
	return
}
