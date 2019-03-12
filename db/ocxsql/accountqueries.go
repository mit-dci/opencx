package ocxsql

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// RegisterUser registers a user
func (db *DB) RegisterUser(pubkey *koblitz.PublicKey, addresses map[match.Asset]string) (err error) {
	// Do all this locking just cause
	// Insert them into the DB
	if err = db.InsertDepositAddresses(pubkey, addresses); err != nil {
		return
	}

	if err = db.InitializeAccountBalances(pubkey); err != nil {
		return
	}

	return
}

// InitializeAccountBalances initializes all database values for an account with username 'username'
func (db *DB) InitializeAccountBalances(pubkey *koblitz.PublicKey) (err error) {

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
		insertBalanceQueries := fmt.Sprintf("INSERT INTO %s VALUES ('%x', %d);", assetString, pubkey.SerializeCompressed(), 0)
		if _, err = tx.Exec(insertBalanceQueries); err != nil {
			return
		}
	}

	return nil
}

// InsertDepositAddresses inserts deposit addresses based on the addressmap you give it
func (db *DB) InsertDepositAddresses(pubkey *koblitz.PublicKey, addressMap map[match.Asset]string) (err error) {
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

	logging.Infof("pubkey length (serialized, compressed): %d", len(pubkey.SerializeCompressed()))

	// use deposit schema
	if _, err = tx.Exec("USE " + db.depositSchema + ";"); err != nil {
		return
	}

	// go through assets
	for _, asset := range db.assetArray {
		// if you found an address in the map
		if addr, found := addressMap[asset]; found {

			logging.Infof("Addr: %s, len: %d", addr, len(addr))
			// insert into db
			insertDepositAddrQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s') ON DUPLICATE KEY UPDATE address='%s';", asset, pubkey.SerializeCompressed(), addr, addr)
			logging.Infof("%s", insertDepositAddrQuery)
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
func (db *DB) UpdateDepositAddresses(ltcAddrFunc func(*koblitz.PublicKey) (string, error), btcAddrFunc func(*koblitz.PublicKey) (string, error), vtcAddrFunc func(*koblitz.PublicKey) (string, error)) (err error) {
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
	for _, asset := range db.assetArray {
		// Find all distinct usernames
		getPubKeysQuery := fmt.Sprintf("SELECT DISTINCT pubkey FROM %s;", asset)
		rows, pubkeyQueryErr := tx.Query(getPubKeysQuery)
		if err = pubkeyQueryErr; err != nil {
			return
		}

		addrPairs := make(map[*koblitz.PublicKey]string)
		// Go through usernames already in db
		for rows.Next() {
			var addr string
			var pubkeyBytes []byte
			if err = rows.Scan(&pubkeyBytes); err != nil {
				return
			}

			var pubkey *koblitz.PublicKey
			if pubkey, err = koblitz.ParsePubKey(pubkeyBytes, koblitz.S256()); err != nil {
				return
			}

			// generate addresses according to chain
			if asset == match.BTCTest {
				if addr, err = btcAddrFunc(pubkey); err != nil {
					return
				}
			} else if asset == match.LTCTest {
				if addr, err = ltcAddrFunc(pubkey); err != nil {
					return
				}
			} else if asset == match.VTCTest {
				if addr, err = vtcAddrFunc(pubkey); err != nil {
					return
				}
			} else {
				err = fmt.Errorf("Tried to update deposit address for unsupported asset")
				return
			}

			// Add usernames and addresses to map
			addrPairs[pubkey] = addr
		}

		// Close the rows so we don't get issues
		if err = rows.Close(); err != nil {
			return
		}

		// go through all the usernames and addresses obtained and update them
		for pubkey, addr := range addrPairs {
			// Actually update the table -- doing this outside the scan so we don't get busy buffer issues
			insertDepositAddrQuery := fmt.Sprintf("UPDATE %s SET address='%s' WHERE pubkey='%x';", asset, addr, pubkey.SerializeCompressed())
			if _, err = tx.Exec(insertDepositAddrQuery); err != nil {
				return
			}
		}
	}

	return
}
