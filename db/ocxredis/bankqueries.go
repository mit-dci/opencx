package ocxredis

// unfinished

// Buyer is a struct containing information needed for a buyer
type Buyer struct {
	username   string
	haveAsset  byte
	// I think 51 bits is enough to store all the satoshis cause 21,000,000 * 10^8 = a 51 bit number
	amountHave int64
}

// Seller is a struct containing information needed for a seller
type Seller struct {
	username   string
	haveAsset  byte
	// I think 51 bits is enough to store all the satoshis cause 21,000,000 * 10^8 = a 51 bit number
	amountHave int64
}

// ExchangeCoins exchanges coins between a buyer and a seller (with a fee of course)
func (db *DB) ExchangeCoins(buyer Buyer, seller Seller) error {
	return nil
}

// InitializeAccount initializes all database values for an account with username 'username'
func (db *DB) InitializeAccount(username string) error {
	return nil
}
