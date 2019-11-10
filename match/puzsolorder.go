package match

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"math/big"
)

// SolutionOrder is an order and modulus that are together.
// This includes an order, and the puzzle modulus factors.
type SolutionOrder struct {
	userOrder AuctionOrder `json:"userorder"`
	p         *big.Int     `json:"p"`
	q         *big.Int     `json:"q"`
}

func NewSolutionOrderFromOrder(aucOrder *AuctionOrder) (solOrder SolutionOrder, err error) {
	if aucOrder == nil {
		err = fmt.Errorf("Cannot create solution order from nil auction order")
		return
	}

	// generate primes p and q
	var rsaPrivKey *rsa.PrivateKey
	if rsaPrivKey, err = rsa.GenerateMultiPrimeKey(rand.Reader, 2, tl.rsaKeyBits); err != nil {
		err = fmt.Errorf("Could not generate primes for RSA: %s", err)
		return
	}
	if len(rsaPrivKey.Primes) != 2 {
		err = fmt.Errorf("For some reason the RSA Privkey has != 2 primes, this should not be the case for RSW, we only need p and q")
		return
	}

	// finally set p, q, and the auction order.
	solOrder.p = new(big.Int).SetBytes(rsaPrivKey.Primes[0].Bytes())
	solOrder.q = new(big.Int).SetBytes(rsaPrivKey.Primes[1].Bytes())
	solOrder.userOrder = *AuctionOrder
	return
}

// TODO: allow for solution order to be turned into puzzle order
