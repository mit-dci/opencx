package match

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// TODO: Next thing to do is to have order signing

// Now this is a good struct, I need hashes & signatures and stuff. The opposite of the dumb asset struct.

// Order is a struct that represents a stored side of a trade
type Order interface {
	Type() string
	Price() (float64, error)
}

// LimitOrder represents a limit order, implementing the order interface
type LimitOrder struct {
	Client      string `json:"username"`
	Side        string `json:"side"`
	TradingPair Pair   `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amount"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64    `json:"price"`
	Timestamp  time.Time `json:"timestamp"`
	OrderID    string    `json:"id"`
}

// IsBuySide returns true if the limit order is buying
func (l *LimitOrder) IsBuySide() bool {
	return l.Side == "buy"
}

// IsSellSide returns true if the limit order is selling
func (l *LimitOrder) IsSellSide() bool {
	return l.Side == "sell"
}

// OppositeSide is a helper to get the opposite side of the order
func (l *LimitOrder) OppositeSide() string {
	if l.IsBuySide() {
		return "sell"
	} else if l.IsSellSide() {
		return "buy"
	}
	return ""
}

// Price gets a float price for the order. This determines how it will get matched. The exchange should figure out if it can take some of the
// pennies off the dollar for things that request a certain amount but the amount they get (according to price and what the other side would be willing
// to give) is less than they officially requested. But tough luck to them we're taking fees anyways
func (l *LimitOrder) Price() (float64, error) {
	if l.AmountWant == 0 {
		return 0, fmt.Errorf("The amount requested in the order is 0, so no price can be calculated. Consider it a donation")
	}
	return float64(l.AmountHave / l.AmountWant), nil
}

// SetID sets an ID for the order, it's going to be different if you call it twice but we're only ever going to call it once
func (l *LimitOrder) SetID() error {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	randNonce := r1.Int63n(math.MaxInt64)
	structBytes := []byte(fmt.Sprintf("%v%x", l, randNonce))

	sha := sha256.New()
	_, err := sha.Write(structBytes)
	if err != nil {
		return err
	}

	l.OrderID = fmt.Sprintf("%x", sha.Sum(nil))
	return nil
}
