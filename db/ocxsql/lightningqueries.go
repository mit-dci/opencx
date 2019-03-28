package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"
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

	updateLightningBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES (%x, %d, %d);", param.Name, pubkey.SerializeCompressed(), qChanID, amount)
	if _, err = tx.Exec(updateLightningBalanceQuery); err != nil {
		return
	}

	return
}
