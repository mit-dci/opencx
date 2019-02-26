package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"github.com/mit-dci/opencx/util"
)

// GetBalance gets the balance of an account
func (db *DB) GetBalance(username string, asset string) (uint64, error) {
	var err error

	// Use the balance schema
	_, err = db.DBHandler.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return 0, fmt.Errorf("Could not use balance schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", asset, username)
	res, err := db.DBHandler.Query(getBalanceQuery)
	// db.IncrementReads()
	if err != nil {
		return 0, fmt.Errorf("Error when getting balance: \n%s", err)
	}

	amount := new(uint64)
	success := res.Next()
	if !success {
		return 0, fmt.Errorf("Database error: username doesn't exist")
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

	coinSchema, err := util.GetSchemaNameFromCoinType(coinType)
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
		insertDepositQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d, %d, %d, '%s');", coinSchema, deposit.Name, deposit.BlockHeightReceived+deposit.Confirmations, deposit.BlockHeightReceived, deposit.Amount, deposit.Txid)
		if _, err = tx.Exec(insertDepositQuery); err != nil {
			return
		}
	}

	areDepositsValidQuery := fmt.Sprintf("SELECT name, amount, txid FROM %s WHERE expectedConfirmHeight=%d;", coinSchema, currentBlockHeight)
	rows, err := tx.Query(areDepositsValidQuery)
	if err != nil {
		return
	}

	// Now we start reflecting the changes for users whose deposits can be filled
	var usernames []string
	var amounts []uint64
	var txids []string

	for rows.Next() {
		var username string
		var amount uint64
		var txid string

		if err = rows.Scan(&username, &amount, &txid); err != nil {
			return
		}

		usernames = append(usernames, username)
		amounts = append(amounts, amount)
		txids = append(txids, txid)
	}
	if err = rows.Close(); err != nil {
		return
	}

	if len(amounts) > 0 {
		if err = db.UpdateBalancesWithinTransaction(usernames, amounts, tx, coinType); err != nil {
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

// UpdateBalance updates a single balance
func (db *DB) UpdateBalance(username string, amount uint64) (err error) {

	tx, err := db.DBHandler.Begin()
	if err != nil {
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

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	currentBalanceQuery := fmt.Sprintf("SELECT balance FROM btc WHERE name='%s';", username)
	rows, err := tx.Query(currentBalanceQuery)
	if err != nil {
		return
	}

	// Get the first result for balance
	rows.Next()
	balance := new(uint64)

	if err = rows.Scan(balance); err != nil {
		return
	}

	if err = rows.Close(); err != nil {
		return
	}

	// Update the balance
	// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
	insertBalanceQuery := fmt.Sprintf("UPDATE btc SET balance='%d' WHERE name='%s';", *balance+amount, username)

	if _, err = tx.Exec(insertBalanceQuery); err != nil {
		return
	}

	return nil
}

// UpdateBalanceWithinTransaction increases the balance of username by amount
func (db *DB) UpdateBalanceWithinTransaction(username string, amount uint64, tx *sql.Tx, coinType *coinparam.Params) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating balances within transaction: \n%s", err)
			return
		}
	}()

	// Another reason why just switching to coinparam stuff is the best option, or at least some sort of coinparam / name struct
	coinSchema, err := util.GetSchemaNameFromCoinType(coinType)
	if err != nil {
		return
	}

	// logging.Infof("Adding %d %s to %s's balance\n", amount, coinSchema, username)
	currentBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", coinSchema, username)
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
		insertBalanceQuery := fmt.Sprintf("UPDATE %s SET balance='%d' WHERE name='%s';", coinSchema, balance+amount, username)
		if _, err = tx.Exec(insertBalanceQuery); err != nil {
			return
		}
	} else {
		// Create the balance
		// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
		insertBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", coinSchema, username, amount)
		if _, err = tx.Exec(insertBalanceQuery); err != nil {
			return
		}
	}

	return nil
}

// UpdateBalancesWithinTransaction updates many balances, uses a transaction for all db stuff.
func (db *DB) UpdateBalancesWithinTransaction(usernames []string, amounts []uint64, tx *sql.Tx, coinType *coinparam.Params) (err error) {

	defer func() {
		if err != nil {
			err = fmt.Errorf("Error updating balances within transaction: \n%s", err)
			return
		}
	}()

	coinSchema, err := util.GetSchemaNameFromCoinType(coinType)
	if err != nil {
		return
	}

	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	for i := range amounts {
		amount := amounts[i]
		username := usernames[i]
		// logging.Infof("Deposit confirmed; Adding %d %s to %s's balance\n", amount, coinSchema, username)
		currentBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", coinSchema, username)
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
			insertBalanceQuery := fmt.Sprintf("UPDATE %s SET balance='%d' WHERE name='%s';", coinSchema, balance+amount, username)
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		} else {
			// Create the balance
			// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
			insertBalanceQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", coinSchema, username, amount)
			if _, err = tx.Exec(insertBalanceQuery); err != nil {
				return
			}
		}
	}

	return nil
}

// GetDepositAddress gets the deposit address of an account
func (db *DB) GetDepositAddress(username string, asset string) (string, error) {
	var err error

	// Use the deposit schema
	_, err = db.DBHandler.Exec("USE " + db.depositSchema + ";")
	if err != nil {
		return "", fmt.Errorf("Could not use deposit schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT address FROM %s WHERE name='%s';", asset, username)
	res, err := db.DBHandler.Query(getBalanceQuery)
	db.IncrementReads()
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
	logging.Debugf("Username %s's deposit address for %s: %s\n", username, asset, *depositAddr)

	err = res.Close()
	if err != nil {
		return "", fmt.Errorf("Error closing deposit result: \n%s", err)
	}

	return *depositAddr, nil

}

// GetDepositName gets the deposit address of an account
func (db *DB) GetDepositName(address string, coinType *coinparam.Params) (string, error) {
	var err error

	asset, err := util.GetSchemaNameFromCoinType(coinType)
	if err != nil {
		return "", fmt.Errorf("Tried to get deposit addresses for %s which isn't a valid asset", asset)
	}
	// Use the deposit schema
	_, err = db.DBHandler.Exec("USE " + db.depositSchema + ";")
	db.IncrementWrites()
	if err != nil {
		return "", fmt.Errorf("Could not use deposit schema: \n%s", err)
	}

	getBalanceQuery := fmt.Sprintf("SELECT address FROM %s WHERE address='%s';", asset, address)
	res, err := db.DBHandler.Query(getBalanceQuery)
	db.IncrementReads()
	if err != nil {
		fmt.Printf("1")
		// return "", fmt.Errorf("Error when getting deposit: \n%s", err)
	}
	getBalanceQuery = fmt.Sprintf("SELECT address FROM %s WHERE address='%s';", asset, address)
	res, err = db.DBHandler.Query(getBalanceQuery)
	db.IncrementReads()
	if err != nil {
		fmt.Printf("2")
		// return "", fmt.Errorf("Error when getting deposit: \n%s", err)
	}
	getBalanceQuery = fmt.Sprintf("SELECT address FROM %s WHERE address='%s';", asset, address)
	res, err = db.DBHandler.Query(getBalanceQuery)
	db.IncrementReads()
	if err != nil {
		fmt.Printf("3\n")
		// return "", fmt.Errorf("Error when getting deposit: \n%s", err)
	}

	if err != nil {
		return "", fmt.Errorf("Error when getting deposit: \n%s", err)
	}

	depositUsername := new(string)
	if res.Next() {
		err = res.Scan(depositUsername)
		if err != nil {
			return "", fmt.Errorf("Error scanning for amount: \n%s", err)
		}
		logging.Debugf("Addr %s's username for %s: %s\n", address, asset, *depositUsername)

	} else {
		return "", nil
	}

	err = res.Close()
	if err != nil {
		return "", fmt.Errorf("Error closing deposit result: \n%s", err)
	}
	return *depositUsername, nil

}

// GetDepositAddressMap returns a map from deposit addresses to names, essentially a set and only to get O(1) access time.
func (db *DB) GetDepositAddressMap(coinType *coinparam.Params) (map[string]string, error) {

	// Use the deposit schema
	_, err := db.DBHandler.Exec("USE " + db.depositSchema + ";")
	if err != nil {
		return nil, fmt.Errorf("Could not use deposit address schema: \n%s", err)
	}

	asset, err := util.GetSchemaNameFromCoinType(coinType)
	if err != nil {
		return nil, fmt.Errorf("Tried to get deposit addresses for %s which isn't a valid asset", asset)
	}
	getBalanceQuery := fmt.Sprintf("SELECT address, name FROM %s;", asset)
	res, err := db.DBHandler.Query(getBalanceQuery)
	if err != nil {
		return nil, fmt.Errorf("Error when getting deposit address: \n%s", err)
	}

	depositAddresses := make(map[string]string)

	for res.Next() {
		var depositAddr string
		var name string
		err = res.Scan(&depositAddr, &name)
		if err != nil {
			return nil, fmt.Errorf("Error scanning for depositAddress: \n%s", err)
		}

		depositAddresses[depositAddr] = name
	}

	err = res.Close()
	if err != nil {
		return nil, fmt.Errorf("Error closing deposit result: \n%s", err)
	}

	return depositAddresses, nil
}

// Withdraw checks that the user has a certain amount of money and removes it if they do
func (db *DB) Withdraw(username string, asset string, amount uint64) (err error) {
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

	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", asset, username)
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
	reduceBalanceQuery := fmt.Sprintf("UPDATE %s SET balance=%d WHERE name='%s'", asset, updatedBalance, username)
	if _, err = tx.Exec(reduceBalanceQuery); err != nil {
		return
	}

	return
}
