package ocxsql

import (
	"fmt"
)

// CreateAccount creates an account
func (db *DB) CreateAccount(username string, password string) (bool, error) {
	err := db.InitializeAccount(username)
	if err != nil {
		return false, fmt.Errorf("Error when trying to create an account: \n%s", err)
	}
	return true, nil
}

// CheckCredentials checks users username and passwords
func (db *DB) CheckCredentials(username string, password string) (bool, error) {
	// TODO later
	return true, nil
}

// InitializeAccount initializes all database values for an account with username 'username'
func (db *DB) InitializeAccount(username string) (err error) {

	// begin the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while initializing accounts: \n%s", err)
	}

	// Setup small custom error msg for trace
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while initializing account: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Use the balance schema
	if _, err = tx.Exec("USE " + db.balanceSchema + ";"); err != nil {
		return
	}

	for _, assetString := range db.assetArray {
		// Insert 0 balance into balance table
		insertBalanceQueries := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", assetString, username, 0)
		if _, err = tx.Exec(insertBalanceQueries); err != nil {
			return
		}
	}

	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		return
	}

	for _, assetString := range db.assetArray {
		var addr string
		if assetString == "btc" {
			if addr, err = db.Keychain.NewAddressBTC(username); err != nil {
				return
			}
		} else if assetString == "ltc" {
			if addr, err = db.Keychain.NewAddressLTC(username); err != nil {
				return
			}
		} else if assetString == "vtc" {
			if addr, err = db.Keychain.NewAddressVTC(username); err != nil {
				return
			}
		}

		insertDepositAddrQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s');", assetString, username, addr)
		if _, err = tx.Exec(insertDepositAddrQuery); err != nil {
			return
		}
	}

	return nil
}
