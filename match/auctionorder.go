package match

import (
	"encoding/binary"
	"fmt"
)

// AuctionOrder represents a limit order, implementing the order interface
type AuctionOrder struct {
	Pubkey      [33]byte `json:"pubkey"`
	Side        string   `json:"side"`
	TradingPair Pair     `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amounthave"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64 `json:"amountwant"`
	// only exists for returning orders back
	OrderbookPrice float64 `json:"orderbookprice"`
	// specify which auction you'd like it to be in
	AuctionID []byte `json:"auctionid"`
}

// IsBuySide returns true if the limit order is buying
func (a *AuctionOrder) IsBuySide() bool {
	return a.Side == "buy"
}

// IsSellSide returns true if the limit order is selling
func (a *AuctionOrder) IsSellSide() bool {
	return a.Side == "sell"
}

// OppositeSide is a helper to get the opposite side of the order
func (a *AuctionOrder) OppositeSide() (sideStr string) {
	if a.IsBuySide() {
		sideStr = "sell"
	} else if a.IsSellSide() {
		sideStr = "buy"
	}
	return
}

// Price gets a float price for the order. This determines how it will get matched. The exchange should figure out if it can take some of the
// pennies off the dollar for things that request a certain amount but the amount they get (according to price and what the other side would be willing
// to give) is less than they officially requested. But tough luck to them we're taking fees anyways
func (a *AuctionOrder) Price() (price float64, err error) {
	if a.AmountWant == 0 {
		err = fmt.Errorf("The amount requested in the order is 0, so no price can be calculated. Consider it a donation")
		return
	}
	if a.IsBuySide() {
		price = float64(a.AmountWant) / float64(a.AmountHave)
		return
	} else if a.IsSellSide() {
		price = float64(a.AmountHave) / float64(a.AmountWant)
	}
	err = fmt.Errorf("Order is not buy or sell, cannot calculate price")
	return
}

// Serialize serializes an order, possible replay attacks here since this is what you're signing?
// but anyways this is the order: pair amountHave amountWant <length side> side
func (a *AuctionOrder) Serialize() (buf []byte) {
	// serializable fields:
	// public key (compressed) [33 bytes]
	// trading pair [2 bytes]
	// amounthave [8 bytes
	// amountwant [8 bytes]
	buf = make([]byte, 33+26+len(a.Side))
	buf = append(buf, a.Pubkey[:]...)
	buf = append(buf, a.TradingPair.Serialize()...)
	binary.LittleEndian.PutUint64(buf, a.AmountHave)
	binary.LittleEndian.PutUint64(buf, a.AmountWant)
	binary.LittleEndian.PutUint64(buf, uint64(len(a.Side)))
	buf = append(buf, []byte(a.Side)...)
	return
}

// SetAmountWant sets the amountwant value of the limit order according to a price
func (a *AuctionOrder) SetAmountWant(price float64) (err error) {
	if price <= 0 {
		err = fmt.Errorf("Price can't be less than or equal to 0")
		return
	}

	if a.IsBuySide() {
		a.AmountWant = uint64(float64(a.AmountHave) * price)
	} else if a.IsSellSide() {
		a.AmountWant = uint64(float64(a.AmountHave) / price)
	} else {
		err = fmt.Errorf("Invalid side for order, must be buy or sell")
		return
	}
	return
}
