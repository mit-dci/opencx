package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"github.com/mit-dci/opencx/util"
)

// GetBalance gets the balance of an account
func (db *DB) GetBalance(pubkey *koblitz.PublicKey, asset string) (uint64, error) {
	var err error

	// Use the balance schema
	_, err = db.DBHandler.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return 0, fmt.Errorf("Could not use balance schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", asset, pubkey.SerializeCompressed())
	res, err := db.DBHandler.Query(getBalanceQuery)
	// db.IncrementReads()
	if err != nil {
		return 0, fmt.Errorf("Error when getting balance: \n%s", err)
	}

	amount := new(uint64)
	success := res.Next()
	if !success {
		return 0, fmt.Errorf("Database error: pubkey doesn't exist")
	}
	err = res.Scan(amount)
	if err != nil {
		return 0, fmt.Errorf("Error scanning for amount: \n%s", err)
	}

	err = res.Close()
	if err != nil {
		return 0, fmt.Errorf("Error closing balance result: \n%s", err)
	}

	return *amount, nil

}

// UpdateDeposits happens when a block is sent in, it updates the deposits table with correct deposits
func (db *DB) UpdateDeposits(deposits []match.Deposit, currentBlockHeight uint64, coinType *coinparam.Params) (err error) {

	coinSchema, err := util.GetSchemaNameFromParam(coinType)
	if err != nil {
		return fmt.Errorf("Error while updating deposits: \n%s", err)
	}

	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while updating deposits: \n%s", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating deposits: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + db.pendingDepositSchema + ";"); err != nil {
		return
	}

	for _, deposit := range deposits {

		// Insert the deposit
		// TODO: replace this name stuff, check that the txid doesn't already exist in deposits. IMPORTANT!!
		insertDepositQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d, %d, %d, '%s');", coinSchema, deposit.Pubkey.SerializeCompressed(), deposit.BlockHeightReceived+deposit.Confirmations, deposit.BlockHeightReceived, deposit.Amount, deposit.Txid)
		if _, err = tx.Exec(insertDepositQuery); err != nil {
			return
		}
	}

	areDepositsValidQuery := fmt.Sprintf("SELECT pubkey, amount, txid FROM %s WHERE expectedConfirmHeight=%d;", coinSchema, currentBlockHeight)
	rows, err := tx.Query(areDepositsValidQuery)
	if err != nil {
		return
	}

	// Now we start reflecting the changes for users whose deposits can be filled
	var pubkeys []*koblitz.PublicKey
	var amounts []uint64
	var txids []string

	for rows.Next() {
		var pubkeyBytes []byte
		var amount uint64
		var txid string

		if err = rows.Scan(&pubkeyBytes, &amount, &txid); err != nil {
			return
		}

		var pubkey *koblitz.PublicKey
		pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256())
		if err != nil {
			return
		}

		pubkeys = append(pubkeys, pubkey)
		amounts = append(amounts, amount)
		txids = append(txids, txid)
	}
	if err = rows.Close(); err != nil {
		return
	}

	if len(amounts) > 0 {
		if err = db.UpdateBalancesWithinTransaction(pubkeys, amounts, tx, coinType); err != nil {
			return
		}
	}

	// HAVE TO GO BACK TO THE DEPOSIT SCHEMA OR ELSE NOTHING WORKS -- dan right after spending 10 minutes on a dumb mistake
	if _, err = tx.Exec("USE " + db.pendingDepositSchema + ";"); err != nil {
		return
	}

	for _, txid := range txids {
		deleteDepositQuery := fmt.Sprintf("DELETE FROM %s WHERE txid='%s';", coinSchema, txid)
		if _, err = tx.Exec(deleteDepositQuery); err != nil {
			return
		}
	}

	return nil
}

// UpdateBalance adds to a single balance
func (db *DB) UpdateBalance(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating balance: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if err = db.UpdateBalanceWithinTransaction(pubkey, amount, tx, param); err != nil {
		return
	}

	return
}

// UpdateBalanceWithinTransaction increases the balance of pubkey by amount
func (db *DB) UpdateBalanceWithinTransaction(pubkey *koblitz.PublicKey, amount uint64, tx *sql.Tx, param *coinparam.Params) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating balances within transaction: \n%s", err)
			return
		}
	}()

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	currentBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", param.Name, pubkey.SerializeCompressed())
	var rows *sql.Rows
	if rows, err = tx.Query(currentBalanceQuery); err != nil {
		return
	}

	// Get the first result for balance
	var balance uint64
	if rows.Next() {

		if err = rows.Scan(&balance); err != nil {
			return
		}
	} else {
		// If we have no results then there's an issue
		err = fmt.Errorf("No balance for pubkey, please register")
		return
	}

	if err = rows.Close(); err != nil {
		return
	}

	insertBalanceQuery := fmt.Sprintf("UPDATE %s SET balance='%d' WHERE pubkey='%x';", param.Name, balance+amount, pubkey.SerializeCompressed())

	if _, err = tx.Exec(insertBalanceQuery); err != nil {
		return
	}

	return
}

// UpdateBalancesWithinTransaction updates many balances, uses a transaction for all db stuff.
func (db *DB) UpdateBalancesWithinTransaction(pubkeys []*koblitz.PublicKey, amounts []uint64, tx *sql.Tx, coinType *coinparam.Params) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating balances within transaction: \n%s", err)
			return
		}
	}()

	coinSchema, err := util.GetSchemaNameFromParam(coinType)
	if err != nil {
		return
	}

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	for i := range amounts {
		amount := amounts[i]
		pubkey := pubkeys[i]
		currentBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", coinSchema, pubkey.SerializeCompressed())
		res, queryErr := tx.Query(currentBalanceQuery)
		if queryErr != nil {
			err = queryErr
			return
		}

		// Get the first result for balance
		if res.Next() {
			var balance uint64

			if err = res.Scan(&balance); err != nil {
				return
			}

			if err = res.Close(); err != nil {
				return
			}

			// Update the balance
			// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
			insertBalanceQuery := fmt.Sprintf("UPDATE %s SET balance='%d' WHERE pubkey='%x';", coinSchema, balance+amount, pubkey.SerializeCompressed())
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		} else {
			// Create the balance
			// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
			insertBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", coinSchema, pubkey.SerializeCompressed(), amount)
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		}
	}

	return nil
}

// GetDepositAddress gets the deposit address of an account
func (db *DB) GetDepositAddress(pubkey *koblitz.PublicKey, asset string) (string, error) {
	var err error

	// Use the deposit schema
	_, err = db.DBHandler.Exec("USE " + db.depositSchema + ";")
	if err != nil {
		return "", fmt.Errorf("Could not use deposit schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT address FROM %s WHERE pubkey='%x';", asset, pubkey.SerializeCompressed())
	res, err := db.DBHandler.Query(getBalanceQuery)
	if err != nil {
		return "", fmt.Errorf("Error when getting deposit: \n%s", err)
	}

	depositAddr := new(string)
	success := res.Next()
	if !success {
		return "", fmt.Errorf("Database error: nothing to scan")
	}
	err = res.Scan(depositAddr)
	if err != nil {
		return "", fmt.Errorf("Error scanning for amount: \n%s", err)
	}
	logging.Debugf("Pubkey %x's deposit address for %s: %s\n", pubkey.SerializeCompressed(), asset, *depositAddr)

	err = res.Close()
	if err != nil {
		return "", fmt.Errorf("Error closing deposit result: \n%s", err)
	}

	return *depositAddr, nil

}

// GetDepositAddressMap returns a map from deposit addresses to pubkeys, essentially a set and only to get O(1) access time.
func (db *DB) GetDepositAddressMap(coinType *coinparam.Params) (depositAddresses map[string]*koblitz.PublicKey, err error) {

	// Use the deposit schema
	if _, err = db.DBHandler.Exec("USE " + db.depositSchema + ";"); err != nil {
		return nil, fmt.Errorf("Could not use deposit address schema: \n%s", err)
	}

	asset, err := util.GetSchemaNameFromParam(coinType)
	if err != nil {
		return nil, fmt.Errorf("Tried to get deposit addresses for %s which isn't a valid asset", asset)
	}
	getBalanceQuery := fmt.Sprintf("SELECT address, pubkey FROM %s;", asset)
	res, err := db.DBHandler.Query(getBalanceQuery)
	if err != nil {
		return nil, fmt.Errorf("Error when getting deposit address: \n%s", err)
	}

	depositAddresses = make(map[string]*koblitz.PublicKey)

	for res.Next() {
		var depositAddr string
		var pubkeyBytes []byte
		err = res.Scan(&depositAddr, &pubkeyBytes)
		if err != nil {
			return nil, fmt.Errorf("Error scanning for depositAddress: \n%s", err)
		}

		var pubkey *koblitz.PublicKey
		if pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
			return
		}

		depositAddresses[depositAddr] = pubkey
	}

	if err = res.Close(); err != nil {
		return nil, fmt.Errorf("Error closing deposit result: \n%s", err)
	}

	return depositAddresses, nil
}

// Withdraw checks that the user has a certain amount of money and removes it if they do
func (db *DB) Withdraw(pubkey *koblitz.PublicKey, asset string, amount uint64) (err error) {
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while withdrawing using database: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Use the balance schema

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", asset, pubkey.SerializeCompressed())
	rows, err := db.DBHandler.Query(getBalanceQuery)
	if err != nil {
		return
	}

	var bal uint64
	if rows.Next() {
		if err = rows.Scan(&bal); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("User not registered, no balance")
		return
	}

	if err = rows.Close(); err != nil {
		return
	}

	if bal < amount {
		err = fmt.Errorf("You do not have enough balance to withdraw this amount")
		return
	}

	updatedBalance := bal - amount
	reduceBalanceQuery := fmt.Sprintf("UPDATE %s SET balance=%d WHERE pubkey='%x'", asset, updatedBalance, pubkey.SerializeCompressed())
	if _, err = tx.Exec(reduceBalanceQuery); err != nil {
		return
	}

	return
}
