package ocxredis

import (
	"crypto/sha256"
	"fmt"
)

// CreateAccount creates an account on the exchange with specified username and password,
// with the USERNAME_PREFIX, username, and the sha256 hash of the password.
func (db *DB) CreateAccount(username string, password string) error {

	passwdHash := sha256.New()
	bytesWritten, err := passwdHash.Write([]byte(password))
	if err != nil {
		return fmt.Errorf("Error hashing password, wrote %d bytes: \n%s", bytesWritten, err)
	}

	prefix := []byte{Account}
	usernameArr := []byte(username)
	qString := string(append(prefix, usernameArr...))
	db.LogPrintf("Querying for username {%s} with query %s\n", username, qString)

	passHashString := fmt.Sprintf("%x", passwdHash.Sum(nil))
	status := db.dbClient.Set(qString, passHashString, 0)
	if status.Err() != nil {
		db.LogErrorf("The DB client had an error creating an account: \n%s", status.Err())
		return status.Err()
	}

	db.LogPrintf("Creating account was a success with username {%s}\n", username)

	return nil
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
	stringcmd := db.dbClient.Get(qString)
	if stringcmd.Err() != nil {
		return false, fmt.Errorf("The DB Client had an error getting the username {%s}: \n%s", username, stringcmd.Err())
	}

	passHashString := fmt.Sprintf("%x", passwdHash.Sum(nil))
	dbresult := stringcmd.String()
	if passHashString != dbresult {
		return false, nil
	}

	db.LogPrintf("User {%s} successfully identified\n", username)
	return true, nil
}
