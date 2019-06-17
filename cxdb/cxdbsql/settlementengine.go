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

	return
}
