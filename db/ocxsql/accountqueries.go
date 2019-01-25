package ocxsql

// this is just here so it can be "implemented"

// CreateAccount creates an account
func(db *DB) CreateAccount(username string, password string) (bool, error) {
	// TODO later
	return true, nil
}

// CheckCredentials checks users username and passwords
func(db *DB) CheckCredentials(username string, password string) (bool, error) {
	// TODO later
	return true, nil
}
