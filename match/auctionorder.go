package match

import (
	"encoding/binary"
	"fmt"

	"github.com/mit-dci/opencx/crypto/timelockencoders"

	"github.com/mit-dci/opencx/crypto"
)

// EncryptedAuctionOrder represents an encrypted Auction Order, so a ciphertext and a puzzle whos solution is a key, and an intended auction.
type EncryptedAuctionOrder struct {
	OrderCiphertext []byte
	OrderPuzzle     crypto.Puzzle
	IntendedAuction [32]byte
}

// OrderPuzzleResult is a struct that is used as the type for a channel so we can atomically
// receive the original encrypted order, decrypted order, and an error
type OrderPuzzleResult struct {
	Encrypted *EncryptedAuctionOrder
	Auction   *AuctionOrder
	Err       error
}

// SolveRC5AuctionOrderAsync solves order puzzles and creates auction orders from them. This should be run in a goroutine.
func (e *EncryptedAuctionOrder) SolveRC5AuctionOrderAsync(puzzleResChan chan *OrderPuzzleResult) {
	var err error
	var result *OrderPuzzleResult
	result.Encrypted = e

	var orderBytes []byte
	if orderBytes, err = timelockencoders.SolvePuzzleRC5(e.OrderCiphertext, e.OrderPuzzle); err != nil {
		result.Err = fmt.Errorf("Error solving RC5 puzzle for auction order: %s", err)
		puzzleResChan <- result
		return
	}

	auctionOrder := new(AuctionOrder)
	if err = auctionOrder.Deserialize(orderBytes); err != nil {
		result.Err = fmt.Errorf("Error deserializing order gotten from puzzle: %s", err)
		puzzleResChan <- result
		return
	}

	result.Auction = auctionOrder
	puzzleResChan <- result

	return
}

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
	// IntendedAuction as the auctionID this should be in. We need this to protect against
	// the exchange withholding an order.
	AuctionID [32]byte `json:"auctionid"`
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
// but anyways this is the order: [33 byte pubkey] pair amountHave amountWant <length side> side [32 byte auctionid]
func (a *AuctionOrder) Serialize() (buf []byte) {
	// serializable fields:
	// public key (compressed) [33 bytes]
	// trading pair [2 bytes]
	// amounthave [8 bytes]
	// amountwant [8 bytes]
	// len side [8 bytes]
	// side [len side]
	// auctionID [32 bytes]
	buf = make([]byte, 32+33+26+len(a.Side))
	buf = append(buf, a.Pubkey[:]...)
	buf = append(buf, a.TradingPair.Serialize()...)
	binary.LittleEndian.PutUint64(buf, a.AmountHave)
	binary.LittleEndian.PutUint64(buf, a.AmountWant)
	binary.LittleEndian.PutUint64(buf, uint64(len(a.Side)))
	buf = append(buf, []byte(a.Side)...)
	buf = append(buf, a.AuctionID[:]...)
	return
}

// Deserialize deserializes an order into the struct ptr it's being called on
func (a *AuctionOrder) Deserialize(data []byte) (err error) {
	// 33 for pubkey, 26 for the rest, 8 for len side, 4 for min side ("sell" is 4 bytes), 32 for auctionID
	if len(data) < 102 {
		err = fmt.Errorf("Auction order cannot be less than 94 bytes: %s", err)
		return
	}

	copy(a.Pubkey[:], data[:33])
	data = data[33:]
	if err = a.TradingPair.Deserialize(data[:2]); err != nil {
		err = fmt.Errorf("Could not deserialize trading pair while deserializing auction order: %s", err)
		return
	}
	data = data[2:]
	a.AmountWant = binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	a.AmountHave = binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	sideLen := binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	a.Side = string(data[:sideLen])
	data = data[sideLen:]
	copy(a.AuctionID[:], data[:32])

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
