package ocxredis

import (
	"crypto/sha256"
	"fmt"
)

// CreateAccount creates an account on the exchange with specified username and password,
// with the USERNAME_PREFIX, username, and the sha256 hash of the password.
func (db *DB) CreateAccount(username string, password string) (bool, error) {


	passwdHash := sha256.New()
	bytesWritten, err := passwdHash.Write([]byte(password))
	if err != nil {
		return false, fmt.Errorf("Error hashing password, wrote %d bytes: \n%s", bytesWritten, err)
	}

	prefix := []byte{Account}
	usernameArr := []byte(username)
	qString := string(append(prefix, usernameArr...))

	db.LogPrintf("Checking that username {%s} does not exist...\n", username)
	intcmd := db.dbClient.Exists(qString)
	if intcmd.Val() != 0 {
		db.LogPrintf("Username {%s} already exists\n", username)
		return false, nil
	}

	db.LogPrintf("Querying for username {%s} with query %s\n", username, qString)

	passHashString := fmt.Sprintf("%x", passwdHash.Sum(nil))
	status := db.dbClient.Set(qString, passHashString, 0)
	if status.Err() != nil {
		db.LogErrorf("The DB client had an error creating an account: \n%s", status.Err())
		return false, status.Err()
	}

	db.LogPrintf("Creating account was a success with username {%s}\n", username)

	return true, nil
}

// CheckCredentials returns a boolean indicating whether or not the credential check succeeded, and an error if an error occurred
func (db *DB) CheckCredentials(username string, password string) (bool, error) {
	passwdHash := sha256.New()
	bytesWritten, err := passwdHash.Write([]byte(password))
	if err != nil {
		return false, fmt.Errorf("Error hashing password, wrote %d bytes: \n%s", bytesWritten, err)
	}

	prefix := []byte{Account}
	usernameArr := []byte(username)
	qString := string(append(prefix, usernameArr...))

	// check that username exists
	db.LogPrintf("Checking that username {%s} exists...\n", username)
	intcmd := db.dbClient.Exists(qString)
	if intcmd.Val() == 0 {
		db.LogPrintf("Username {%s} doesn't exist\n", username)
		return false, nil
	}

	// get passwordhash
	stringcmd := db.dbClient.Get(qString)
	dbresult, err := stringcmd.Result()
	if err != nil {
		return false, fmt.Errorf("The DB Client had an error getting the username {%s}: \n%s", username, err)
	}

	// compare with hash of password arg
	passHashString := fmt.Sprintf("%x", passwdHash.Sum(nil))
	if passHashString != dbresult {
		return false, nil
	}

	db.LogPrintf("User {%s} successfully identified\n", username)
	return true, nil
}
