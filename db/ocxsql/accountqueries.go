package ocxsql

import (
	"fmt"

	"github.com/mit-dci/opencx/logging"
)

// InitializeAccountBalances initializes all database values for an account with username 'username'
func (db *DB) InitializeAccountBalances(username string) (err error) {

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

	return nil
}

// InsertDepositAddresses inserts deposit addresses based on the addressmap you give it
func (db *DB) InsertDepositAddresses(username string, addressMap map[string]string) (err error) {
	// begin the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while inserting addresses: \n%s", err)
	}

	// Setup small custom error msg for trace
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while inserting addresses: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use deposit schema
	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		return
	}

	// go through assets
	for _, assetString := range db.assetArray {
		// if you found an address in the map
		if addr, found := addressMap[assetString]; found {
			// insert into db
			insertDepositAddrQuery := fmt.Sprintf("REPLACE INTO %s VALUES ('%s', '%s');", assetString, username, addr)
			if _, err = tx.Exec(insertDepositAddrQuery); err != nil {
				return
			}
		} else {
			err = fmt.Errorf("Could not find asset while creating addresses")
			return
		}
	}

	return
}

// UpdateDepositAddresses updates deposit addresses based on the usernames in the table. Doing function stuff so it looks weird
func (db *DB) UpdateDepositAddresses(ltcAddrFunc func(string) (string, error), btcAddrFunc func(string) (string, error), vtcAddrFunc func(string) (string, error)) (err error) {
	// begin the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while updating addresses: \n%s", err)
	}

	// Setup small custom error msg for trace
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating addresses: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Use deposit schema
	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		return
	}

	// Go through assets (btc, ltc, vtc...)
	for _, assetString := range db.assetArray {
		// Find all distinct usernames
		getUsernamesQuery := fmt.Sprintf("SELECT DISTINCT name FROM %s;", assetString)
		rows, usernameErr := tx.Query(getUsernamesQuery)
		if err = usernameErr; err != nil {
			return
		}

		addrPairs := make(map[string]string)
		// Go through usernames already in db
		for rows.Next() {
			var addr string
			var username string
			if err = rows.Scan(&username); err != nil {
				return
			}

			// generate addresses according to chain
			if assetString == "btc" {
				if addr, err = btcAddrFunc(username); err != nil {
					return
				}
			} else if assetString == "ltc" {
				if addr, err = ltcAddrFunc(username); err != nil {
					return
				}
			} else if assetString == "vtc" {
				if addr, err = vtcAddrFunc(username); err != nil {
					return
				}
			} else {
				err = fmt.Errorf("Tried to update deposit address for unsupported asset")
				return
			}

			logging.Infof("Got address for %s on %s chain: %s\n", username, assetString, addr)

			// Add usernames and addresses to map
			addrPairs[username] = addr
		}

		// Close the rows so we don't get issues
		if err = rows.Close(); err != nil {
			return
		}

		// go through all the usernames and addresses obtained and update them
		for addr, username := range addrPairs {

			// Actually update the table -- doing this outside the scan so we don't get busy buffer issues
			insertDepositAddrQuery := fmt.Sprintf("UPDATE %s SET address='%s' WHERE name='%s';", assetString, addr, username)
			if _, err = tx.Exec(insertDepositAddrQuery); err != nil {
				return
			}
		}
	}

	return
}
