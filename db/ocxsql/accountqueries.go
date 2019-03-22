package ocxsql

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"

	"github.com/mit-dci/opencx/logging"
)

// RegisterUser registers a user
func (db *DB) RegisterUser(pubkey *koblitz.PublicKey, addresses map[*coinparam.Params]string) (err error) {
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
func (db *DB) InsertDepositAddresses(pubkey *koblitz.PublicKey, addressMap map[*coinparam.Params]string) (err error) {
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

	for chain, addr := range addressMap {
		logging.Infof("Addr: %s, len: %d", addr, len(addr))
		// insert into db
		insertDepositAddrQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s') ON DUPLICATE KEY UPDATE address='%s';", chain.Name, pubkey.SerializeCompressed(), addr, addr)
		logging.Infof("%s", insertDepositAddrQuery)
		if _, err = tx.Exec(insertDepositAddrQuery); err != nil {
			return
		}
	}

	return
}
