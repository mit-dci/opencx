package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
)

// LightningDeposit updates the fund balance for a specific pubkey.
func (db *DB) LightningDeposit(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params, qChanID uint32) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while updating funds: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating funds: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.lightningbalanceSchema + ";"); err != nil {
		return
	}

	updateLightningBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d, %d);", param.Name, pubkey.SerializeCompressed(), qChanID, amount)
	logging.Infof("lightning query: %s", updateLightningBalanceQuery)
	if _, err = tx.Exec(updateLightningBalanceQuery); err != nil {
		return
	}

	return
}

// LightningWithdraw will remove money from a user's tracked lightning balance or return an error.
func (db *DB) LightningWithdraw(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {
	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while updating funds: \n%s", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating funds: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.lightningbalanceSchema + ";"); err != nil {
		return
	}

	updateLightningBalanceQuery := fmt.Sprintf("SELECT amount FROM %s WHERE pubkey='%s';", param.Name, pubkey.SerializeCompressed())
	logging.Infof("lightning query: %s", updateLightningBalanceQuery)
	var res *sql.Rows
	if res, err = tx.Query(updateLightningBalanceQuery); err != nil {
		return
	}

	var totalAmount uint64

	var amountGot uint64
	for res.Next() {
		if err = res.Scan(&amountGot); err != nil {
			err = fmt.Errorf("Error scanning for amount: \n%s", err)
			return
		}

		totalAmount += amountGot
	}

	if err = res.Close(); err != nil {
		err = fmt.Errorf("Error closing balance result: \n%s", err)
		return
	}

	return
}
