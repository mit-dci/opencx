package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// GetBalance gets the balance of an account
func (db *DB) GetBalance(pubkey *koblitz.PublicKey, param *coinparam.Params) (uint64, error) {
	var err error

	// Use the balance schema
	_, err = db.DBHandler.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return 0, fmt.Errorf("Could not use balance schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", param.Name, pubkey.SerializeCompressed())
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
		insertDepositQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d, %d, %d, '%s');", coinType.Name, deposit.Pubkey.SerializeCompressed(), deposit.BlockHeightReceived+deposit.Confirmations, deposit.BlockHeightReceived, deposit.Amount, deposit.Txid)
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

		// so res.Scan does something weird because I'm not using prepared statements. I would prefer to be using a byte array to scan into, since the pubkey is a varbinary?
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
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
		deleteDepositQuery := fmt.Sprintf("DELETE FROM %s WHERE txid='%s';", coinType.Name, txid)
		if _, err = tx.Exec(deleteDepositQuery); err != nil {
			return
		}
	}

	return nil
}

// AddToBalance adds to a single balance
func (db *DB) AddToBalance(pubkey *koblitz.PublicKey, amount uint64, param *coinparam.Params) (err error) {

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

	if err = db.AddToBalanceWithinTransaction(pubkey, amount, tx, param); err != nil {
		return
	}

	return
}

// AddToBalanceWithinTransaction increases the balance of pubkey by amount
func (db *DB) AddToBalanceWithinTransaction(pubkey *koblitz.PublicKey, amount uint64, tx *sql.Tx, param *coinparam.Params) (err error) {
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

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	for i := range amounts {
		amount := amounts[i]
		pubkey := pubkeys[i]
		currentBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", coinType.Name, pubkey.SerializeCompressed())
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
			insertBalanceQuery := fmt.Sprintf("UPDATE %s SET balance='%d' WHERE pubkey='%x';", coinType.Name, balance+amount, pubkey.SerializeCompressed())
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		} else {
			// Create the balance
			// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
			insertBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", coinType.Name, pubkey.SerializeCompressed(), amount)
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		}
	}

	return nil
}

// GetDepositAddress gets the deposit address of an account
func (db *DB) GetDepositAddress(pubkey *koblitz.PublicKey, asset string) (depositAddr string, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}
	// this defer is pushed onto the stack first
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting deposit address: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	// Use the deposit schema
	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use deposit schema: \n%s", err)
		return
	}

	getBalanceQuery := fmt.Sprintf("SELECT address FROM %s WHERE pubkey='%x';", asset, pubkey.SerializeCompressed())
	var rows *sql.Rows
	if rows, err = tx.Query(getBalanceQuery); err != nil {
		err = fmt.Errorf("Error when getting deposit: \n%s", err)
		return
	}

	if rows.Next() {

		if err = rows.Scan(&depositAddr); err != nil {
			err = fmt.Errorf("Error scanning for deposit address: \n%s", err)
			return
		}

		logging.Debugf("Pubkey %x's deposit address for %s: %s\n", pubkey.SerializeCompressed(), asset, depositAddr)

	} else {
		err = fmt.Errorf("Cannot find deposit address. Make sure you've registered")
		return
	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing deposit result: \n%s", err)
		return
	}

	return
}

// GetDepositAddressMap returns a map from deposit addresses to pubkeys, essentially a set and only to get O(1) access time.
func (db *DB) GetDepositAddressMap(coinType *coinparam.Params) (depositAddresses map[string]*koblitz.PublicKey, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}
	// this defer is pushed onto the stack first
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting deposit addr map: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	// Use the deposit schema
	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use deposit address schema: \n%s", err)
		return
	}

	var getBalanceStmt *sql.Stmt
	if getBalanceStmt, err = tx.Prepare(fmt.Sprintf("SELECT pubkey, address FROM %s;", coinType.Name)); err != nil {
		return
	}

	// defer stmt.close so it still is executed but doesn't interfere with closing other stuff if it errors out
	// this defer is pushed onto the defer stack second, so we close out the statement, then close out the tx.
	defer func() {
		if err != nil {
			if newErr := getBalanceStmt.Close(); newErr != nil {
				err = fmt.Errorf("Error with closing getbalancestmt: \n%s\nAdditional errors: \n%s", newErr, err)
			}
		} else {
			if newErr := getBalanceStmt.Close(); newErr != nil {
				err = fmt.Errorf("Error with closing getbalancestmt: \n%s", newErr)
			}
		}
	}()

	var rows *sql.Rows
	if rows, err = getBalanceStmt.Query(); err != nil {
		err = fmt.Errorf("Error when getting deposit address: \n%s", err)
		return
	}

	depositAddresses = make(map[string]*koblitz.PublicKey)

	for rows.Next() {
		var depositAddr string
		var pubkeyBytes []byte
		if err = rows.Scan(&pubkeyBytes, &depositAddr); err != nil {
			err = fmt.Errorf("Error scanning for depositAddrowss: \n%s", err)
			return
		}

		// so rows.Scan does something weird because I'm not using prepared statements. I would prefer to be using a byte array to scan into, since the pubkey is a varbinary?
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			return
		}

		var pubkey *koblitz.PublicKey
		if pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
			err = fmt.Errorf("Error parsing pubkey %x for curve S256: \n%s", pubkeyBytes, err)
			return
		}

		depositAddresses[depositAddr] = pubkey
	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing deposit rows: \n%s", err)
		return
	}

	return
}

// Withdraw checks that the user has a certain amount of money and removes it if they do
func (db *DB) Withdraw(pubkey *koblitz.PublicKey, asset *coinparam.Params, amount uint64) (err error) {
	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while withdrawing %s using database: \n%s", asset.Name, err)
			return
		}
		err = tx.Commit()
	}()

	// Use the balance schema

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		err = fmt.Errorf("Error trying to use database: \n%s", err)
		return
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE pubkey='%x';", asset.Name, pubkey.SerializeCompressed())
	var rows *sql.Rows
	if rows, err = tx.Query(getBalanceQuery); err != nil {
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
	reduceBalanceQuery := fmt.Sprintf("UPDATE %s SET balance=%d WHERE pubkey='%x'", asset.Name, updatedBalance, pubkey.SerializeCompressed())
	if _, err = tx.Exec(reduceBalanceQuery); err != nil {
		return
	}

	return
}
