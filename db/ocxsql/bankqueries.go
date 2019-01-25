package ocxsql

import (
	"github.com/mit-dci/opencx/match"
)

// ExchangeCoins exchanges coins between a buyer and a seller (with a fee of course)
func (db *DB) ExchangeCoins(buyOrder *match.Order, sellOrder *match.Order) error {
	// check balances
	// if balances check out then make the trade, update balances

	return nil
}

// InitializeAccount initializes all database values for an account with username 'username'
func (db *DB) InitializeAccount(username string) error {
	// Balances table, username is one column and balances are all the other columns
	// SELECT username FROM balances...
	// tx, err := db.Begin()
	return nil
}
