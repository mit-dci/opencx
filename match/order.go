package match

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Order is a struct that represents a stored side of a trade
type Order interface {
	Type() string
	Price() (float64, error)
}

// LimitOrder represents a limit order, implementing the order interface
type LimitOrder struct {
	Pubkey      [33]byte `json:"pubkey"`
	Side        string   `json:"side"`
	TradingPair Pair     `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amounthave"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64    `json:"amountwant"`
	Timestamp  time.Time `json:"timestamp"`
	OrderID    string    `json:"id"`
	// only exists for returning orders back
	OrderbookPrice float64 `json:"orderbookprice"`
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
func (l *LimitOrder) OppositeSide() (sideStr string) {
	if l.IsBuySide() {
		sideStr = "sell"
	} else if l.IsSellSide() {
		sideStr = "buy"
	}
	return
}

// Price gets a float price for the order. This determines how it will get matched. The exchange should figure out if it can take some of the
// pennies off the dollar for things that request a certain amount but the amount they get (according to price and what the other side would be willing
// to give) is less than they officially requested. But tough luck to them we're taking fees anyways
func (l *LimitOrder) Price() (float64, error) {
	if l.AmountWant == 0 {
		return 0, fmt.Errorf("The amount requested in the order is 0, so no price can be calculated. Consider it a donation")
	}
	if l.IsBuySide() {
		return float64(l.AmountWant) / float64(l.AmountHave), nil
	} else if l.IsSellSide() {
		return float64(l.AmountHave) / float64(l.AmountWant), nil
	}
	return 0, fmt.Errorf("Order is not buy or sell, cannot calculate price")
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

// Serialize serializes an order, possible replay attacks here since this is what you're signing?
// but anyways this is the order: pair amountHave amountWant <length side> side
func (l *LimitOrder) Serialize() (buf []byte) {
	// serializable fields:
	// public key (compressed) [33 bytes]
	// trading pair [2 bytes]
	// amounthave [8 bytes
	// amountwant [8 bytes]
	buf = make([]byte, 33+26+len(l.Side))
	buf = append(buf, l.Pubkey[:]...)
	buf = append(buf, l.TradingPair.Serialize()...)
	binary.LittleEndian.PutUint64(buf, l.AmountHave)
	binary.LittleEndian.PutUint64(buf, l.AmountWant)
	binary.LittleEndian.PutUint64(buf, uint64(len(l.Side)))
	buf = append(buf, []byte(l.Side)...)
	return
}

// SetAmountWant sets the amountwant value of the limit order according to a price
func (l *LimitOrder) SetAmountWant(price float64) error {
	if price <= 0 {
		return fmt.Errorf("Price can't be less than or equal to 0")
	}

	// Rules for all of this amountHave / amountWant confusing stuff because I'm bad at naming variables:
	// Say the market is called BTC/USD
	// This means that if your order is 'buying' then you already have USD and you're looking to buy BTC
	// If you're buying your amountHave will be in units of USD, and your amountWant will be in units of BTC
	// On the other hand, if your side is 'selling' then you already have BTC and you're looking to sell it for USD
	// If you're selling then your amountHave will be in units of BTC and your amountWant will be in units of USD
	// THIS is all different from the assetHave and assetWant in the trading pair
	// in the trading pair, the assetWant is supposed to be the primary asset, so that is BTC in this case
	// the assetWant / have is all from a buyer's point of view in the trading pair struct.
	// just putting this here because this is also where crucial price code is, because price means the SAME thing
	// across a buyer and seller, while everything else means different things.
	if l.IsBuySide() {
		l.AmountWant = uint64(float64(l.AmountHave) * price)
	} else if l.IsSellSide() {
		l.AmountWant = uint64(float64(l.AmountHave) / price)
	} else {
		return fmt.Errorf("Invalid side for order, must be buy or sell")
	}
	return nil
}
