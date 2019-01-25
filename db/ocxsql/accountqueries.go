package ocxsql

import (
	"fmt"
)

// CreateAccount creates an account
func(db *DB) CreateAccount(username string, password string) (bool, error) {
	err := db.InitializeAccount(username)
	if err != nil {
		return false, fmt.Errorf("Error when trying to create an account: \n%s", err)
	}
	return true, nil
}

// CheckCredentials checks users username and passwords
func(db *DB) CheckCredentials(username string, password string) (bool, error) {
	// TODO later
	return true, nil
}
