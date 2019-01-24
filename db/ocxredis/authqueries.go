package ocxredis

import (
	"time"
	"fmt"
	"math/rand"
	"crypto/sha256"
)

// This is here for reminder, use the expiration parameter in the Set() method for creating tokens!!

// CreateStoreToken creates a token and stores it
func(db *DB) CreateStoreToken(username string) ([]byte, error) {


	prefix := []byte{Token}
	usernameArr := []byte(username)
	qString := string(append(prefix, usernameArr...))

	randHash := sha256.New()

	// Just make a really big random number that has no security features other than the fact that it's really big
	largeRandString := fmt.Sprintf("%x", rand.Uint64()) + fmt.Sprintf("%x", rand.Uint64()) + fmt.Sprintf("%x", rand.Uint64()) + fmt.Sprintf("%x", rand.Uint64())
	randHash.Write([]byte(largeRandString))

	tokenString := fmt.Sprintf("%x", randHash.Sum(nil))

	status := db.dbClient.Set(qString, tokenString, 20 * time.Second)
	err := status.Err()
	if err != nil {
		db.LogErrorf("The DB client had an error creating a token: \n%s", err)
		return nil, err
	}

	return []byte(tokenString), nil
}

// TODO: will probably need to add user variables in the Server so that gets passed along with stuff

// CheckToken checks that the token checks out with the username
func(db *DB) CheckToken(username string, token []byte) (bool, error) {

	prefix := []byte{Token}
	usernameArr := []byte(username)
	qString := string(append(prefix, usernameArr...))


	// Check that username exists
	db.LogPrintf("Checking that token for username {%s} exists...\n", username)
	intcmd := db.dbClient.Exists(qString)
	if intcmd.Val() == 0 {
		db.LogPrintf("Token for Username {%s} doesn't exist\n", username)
		return false, nil
	}

	db.LogPrintf("Checking token provided by client...\n", username)
	status := db.dbClient.Get(qString)
	res, err := status.Result()
	if err != nil {
		db.LogErrorf("The DB client had an error creating an account: \n%s", err)
		return false, err
	}

	tokenString := fmt.Sprintf("%s", token)
	if res != tokenString {
		db.LogPrintf("Token check failed for %s\n", username)
		return false, nil
	}

	return true, nil
}
