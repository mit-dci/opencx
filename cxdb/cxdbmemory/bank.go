package cxdbmemory

import (
	"fmt"

	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"
)

// RegisterUser takes in a pubkey, and a map of asset to addresses for the pubkey. It inserts the necessary information in databases to register the pubkey.
func (db *CXDBMemory) RegisterUser(pubkey *koblitz.PublicKey, addressMap map[*coinparam.Params]string) (err error) {

	db.balancesMtx.Lock()
	// We don't care about addresses because deposit stuff is not implemented yet and this shouldn't be used for production
	for coin := range addressMap {

		var pkpair pubkeyCoinPair
		copy(pkpair.pubkey[:], pubkey.SerializeCompressed())
		pkpair.coin = coin

		db.balances[&pkpair] = 0
	}
	db.balancesMtx.Unlock()

	return
}

// GetBalance gets the balance for a pubkey and an asset.
func (db *CXDBMemory) GetBalance(pubkey *koblitz.PublicKey, coin *coinparam.Params) (amount uint64, err error) {

	var pkpair pubkeyCoinPair
	copy(pkpair.pubkey[:], pubkey.SerializeCompressed())
	pkpair.coin = coin

	db.balancesMtx.Lock()
	var found bool
	if amount, found = db.balances[&pkpair]; !found {
		db.balancesMtx.Unlock()
		err = fmt.Errorf("Could not find balance, register please")
		return
	}
	db.balancesMtx.Unlock()

	return
}

// AddToBalance adds to the balance of a user
func (db *CXDBMemory) AddToBalance(pubkey *koblitz.PublicKey, amount uint64, coin *coinparam.Params) (err error) {

	var pkpair pubkeyCoinPair
	copy(pkpair.pubkey[:], pubkey.SerializeCompressed())
	pkpair.coin = coin

	db.balancesMtx.Lock()
	var found bool
	var oldAmt uint64
	if oldAmt, found = db.balances[&pkpair]; !found {
		db.balancesMtx.Unlock()
		err = fmt.Errorf("Could not find balance, register please")
		return
	}
	db.balances[&pkpair] = oldAmt + amount
	db.balancesMtx.Unlock()

	return
}

// Withdraw checks the user's balance against the amount and if valid, reduces the balance by that amount.
func (db *CXDBMemory) Withdraw(pubkey *koblitz.PublicKey, coin *coinparam.Params, amount uint64) (err error) {

	var pkpair pubkeyCoinPair
	copy(pkpair.pubkey[:], pubkey.SerializeCompressed())
	pkpair.coin = coin

	db.balancesMtx.Lock()
	var found bool
	var oldAmt uint64
	if oldAmt, found = db.balances[&pkpair]; !found {
		db.balancesMtx.Unlock()
		err = fmt.Errorf("Could not find balance, register please")
		return
	}

	if oldAmt < amount {
		err = fmt.Errorf("You do not have enough balance to withdraw this amount")
		return
	}

	db.balances[&pkpair] = oldAmt - amount
	db.balancesMtx.Unlock()

	return
}
