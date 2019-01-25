package ocxsql

// CreateStoreToken creates a token for username and stores it for a certain amount of time
func(db *DB) CreateStoreToken(username string) ([]byte, error) {
	// TODO later
	return nil, nil
}

// CheckToken checks the token assigned to a user
func(db *DB) CheckToken(username string, token []byte) (bool, error) {
	// TODO maybe never if I use signed stuff
	return true, nil
}
