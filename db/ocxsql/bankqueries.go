package ocxsql

import (
	"fmt"

	"github.com/mit-dci/lit/coinparam"

	// "database/sql"

	"github.com/mit-dci/opencx/match"
	// mysql is just the driver, always interact with database/sql api
	// _ "github.com/go-sql-driver/mysql"
)

// ExchangeCoins exchanges coins between a buyer and a seller (with a fee of course)
func (db *DB) ExchangeCoins(buyOrder *match.Order, sellOrder *match.Order) error {
	// check balances
	// if balances check out then make the trade, update balances

	return nil
}

// InitializeAccount initializes all database values for an account with username 'username'
func (db *DB) InitializeAccount(username string) error {
	var err error

	// Use the balance schema
	_, err = db.DBHandler.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return fmt.Errorf("Could not use balance schema: \n%s", err)
	}

	// begin the transaction
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while initializing accounts: \n%s", err)
	}

	for _, assetString := range db.assetArray {
		insertBalanceQueries := fmt.Sprintf("INSERT INTO %s VALUES ('%s', %d);", assetString, username, 0)
		_, err := tx.Exec(insertBalanceQueries)
		if err != nil {
			return fmt.Errorf("Error creating initial balance: \n%s", err)
		}
	}

	// for _, assetString := range db.assetArray {
	// 	// insertDepositAddrQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s')", assetString, username)
	// }

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error committing transaction: \n%s", err)
	}
	return nil
}

// GetBalance gets the balance of an account
func (db *DB) GetBalance(username string, asset string) (uint64, error) {
	var err error

	// Use the balance schema
	_, err = db.DBHandler.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return 0, fmt.Errorf("Could not use balance schema: \n%s", err)
	}

	// Check if the asset exists
	validTable := false
	for _, elem := range db.assetArray {
		if asset == elem {
			validTable = true
		}
	}

	if !validTable {
		return 0, fmt.Errorf("User %s tried to get balance for %s which isn't a valid asset", username, asset)
	}
	getBalanceQuery := fmt.Sprintf("SELECT balance FROM %s WHERE name='%s';", asset, username)
	res, err := db.DBHandler.Query(getBalanceQuery)
	if err != nil {
		return 0, fmt.Errorf("Error when getting balance: \n%s", err)
	}

	amount := new(uint64)
	success := res.Next()
	if !success {
		return 0, fmt.Errorf("Database error: nothing to scan")
	}
	err = res.Scan(amount)
	if err != nil {
		return 0, fmt.Errorf("Error scanning for amount: \n%s", err)
	}
	db.LogPrintf("%s balance for %s: %d\n", username, asset, *amount)

	err = res.Close()
	if err != nil {
		return 0, fmt.Errorf("Error closing balance result: \n%s", err)
	}

	return *amount, nil

}

// GetBalances gets all balances of an account
func (db *DB) GetBalances(username string) error {
	return nil
}

// UpdateDeposits happens when a block is sent in, it updates the deposits table with correct deposits
func (db *DB) UpdateDeposits(amounts []uint64, coinType coinparam.Params) error {

	db.LogPrintf("Updating deposits for a block!\n")
	tx, err := db.DBHandler.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction while updating deposits: \n%s", err)
	}

	_, err = tx.Exec("USE " + db.balanceSchema + ";")
	if err != nil {
		return fmt.Errorf("Error using balance schema: \n%s", err)
	}

	for _, amt := range amounts {
		currentBalanceQuery := fmt.Sprintf("SELECT balance FROM btc WHERE name='%s';", "dan")
		res, err := tx.Query(currentBalanceQuery)
		if err != nil {
			return fmt.Errorf("Error when getting the current balance: \n%s", err)
		}

		// Get the first result for balance
		res.Next()
		balance := new(uint64)
		err = res.Scan(balance)
		if err != nil {
			return fmt.Errorf("Error when scanning for name: \n%s", err)
		}
		err = res.Close()
		if err != nil {
			return fmt.Errorf("Error when closing rows: \n%s", err)
		}

		// Update the balance
		// TODO: replace this name stuff, check that the name doesn't already exist on register. IMPORTANT!!
		insertBalanceQuery := fmt.Sprintf("UPDATE btc SET balance='%d' WHERE name='%s';", *balance+amt, "dan")
		_, err = tx.Exec(insertBalanceQuery)
		if err != nil {
			return fmt.Errorf("Error inserting into balance table: \n%s", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Error committing transaction: \n%s", err)
	}
	db.LogPrintf("Done updating deposits for this block!\n")

	return nil
}
