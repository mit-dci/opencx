package match

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/mit-dci/opencx/crypto/timelockencoders"
)

// OrderPuzzleResult is a struct that is used as the type for a channel so we can atomically
// receive the original encrypted order, decrypted order, and an error
type OrderPuzzleResult struct {
	Encrypted *EncryptedAuctionOrder
	Auction   *AuctionOrder
	Err       error
}

// AuctionOrder represents a batch order
type AuctionOrder struct {
	Pubkey      [33]byte `json:"pubkey"`
	Side        string   `json:"side"`
	TradingPair Pair     `json:"pair"`
	// amount of assetHave the user would like to trade
	AmountHave uint64 `json:"amounthave"`
	// amount of assetWant the user wants for their assetHave
	AmountWant uint64 `json:"amountwant"`
	// IntendedAuction as the auctionID this should be in. We need this to protect against
	// the exchange withholding an order.
	AuctionID [32]byte `json:"auctionid"`
	// 2 byte nonce (So there can be max 2^16 of the same-looking orders by the same pubkey in the same batch)
	// This is used to protect against the exchange trying to replay a bunch of orders
	Nonce     [2]byte `json:"nonce"`
	Signature []byte  `json:"signature"`
}

// TODO: create an order ID method that hashes the Nonce and Signature? People should be able to verify the signature whenever, even if partially filled.

// TurnIntoEncryptedOrder creates a puzzle for this auction order given the time. We make no assumptions about whether or not the order is signed.
func (a *AuctionOrder) TurnIntoEncryptedOrder(t uint64) (encrypted *EncryptedAuctionOrder, err error) {
	encrypted = new(EncryptedAuctionOrder)
	if encrypted.OrderCiphertext, encrypted.OrderPuzzle, err = timelockencoders.CreateRSW2048A2PuzzleRC5(t, a.Serialize()); err != nil {
		err = fmt.Errorf("Error creating puzzle from auction order: %s", err)
		return
	}
	// make sure they match
	encrypted.IntendedAuction = a.AuctionID
	return
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
	if a.AmountWant == 0 || a.AmountHave == 0 {
		err = fmt.Errorf("The amount requested in the order is 0, so no price can be calculated")
		return
	}
	price = float64(a.AmountWant) / float64(a.AmountHave)
	return
}

// GenerateOrderFill creates an execution that will fill an order (AmountHave at the end is 0) and provides an order and settlement execution.
// This does not assume anything about the price of the order, as we can't infer what price the order was
// placed at.
// TODO: Figure out whether or not these should be pointers
func (a *AuctionOrder) GenerateOrderFill(orderID []byte, execPrice float64) (orderExec OrderExecution, setExec SettlementExecution, err error) {

	if a.AmountHave == 0 {
		err = fmt.Errorf("Error generating order fill: empty order, the AmountHave cannot be 0")
		return
	}

	if execPrice == float64(0) {
		err = fmt.Errorf("Error generating order fill: price cannot be zero")
		return
	}
	// So we want to make sure that we're filling this, so first calculate the amountWantToFill and make sure it equals amountHave
	var amountToDebit uint64
	var debitAsset Asset
	var creditAsset Asset
	if a.IsBuySide() {
		debitAsset = a.TradingPair.AssetWant
		creditAsset = a.TradingPair.AssetHave
	} else if a.IsSellSide() {
		debitAsset = a.TradingPair.AssetHave
		creditAsset = a.TradingPair.AssetWant
	} else {
		err = fmt.Errorf("Error generating order fill from price, order is not buy or sell side, it's %s side", a.Side)
		return
	}
	amountToDebit = uint64(float64(a.AmountHave) * execPrice)

	// IMPORTANT! These lines:
	// > OrderID: make([]byte, len(orderID),
	// > copy(execution.OrderID, orderID)
	// are EXTREMELY IMPORTANT because there's pretty much no other way to copy
	// bytes. This was a really annoying issue to debug
	// Finally generate execution
	orderExec = OrderExecution{
		OrderID:       make([]byte, len(orderID)),
		NewAmountWant: 0,
		NewAmountHave: 0,
		Filled:        true,
	}
	setExec = SettlementExecution{
		Debited: Entry{
			Amount: amountToDebit,
			Asset:  debitAsset,
		},
		Credited: Entry{
			Amount: a.AmountHave,
			Asset:  creditAsset,
		},
	}
	copy(setExec.Pubkey[:], a.Pubkey[:])
	copy(orderExec.OrderID, orderID)
	return
}

// GenerateExecutionFromPrice generates a trade execution from a price and an amount to fill. This is intended to be
// used by the matching engine when a price is determined for this order to execute at.
// amountToFill refers to the amount of AssetWant that can be filled. So the other side's "AmountHave" can be passed
// in as a parameter. The order ID will be filled in, as it's being passed as a parameter.
// This returns a fillRemainder, which is the amount that is left over from amountToFill after
// filling orderID at execPrice and amountToFill
func (a *AuctionOrder) GenerateExecutionFromPrice(orderID []byte, execPrice float64, amountToFill uint64) (orderExec OrderExecution, setExec SettlementExecution, fillRemainder uint64, err error) {
	// If it's a buy side, AmountWant is assetWant, and AmountHave is assetHave - but price is something different, price is want/have.
	// So to convert from amountWant (amountToFill) to amountHave we need to multiple amountToFill by 1/execPrice
	var amountWantToFill uint64
	var debitAsset Asset
	var creditAsset Asset
	if a.IsBuySide() {
		debitAsset = a.TradingPair.AssetHave
		creditAsset = a.TradingPair.AssetWant
	} else if a.IsSellSide() {
		debitAsset = a.TradingPair.AssetWant
		creditAsset = a.TradingPair.AssetHave
	} else {
		err = fmt.Errorf("Error generating execution from price, order is not buy or sell side, it's %s side", a.Side)
		return
	}
	amountWantToFill = uint64(float64(amountToFill) * execPrice)

	// Now that we have this value, we'll generate the execution.
	// What should we do if the amountWantToFill is greater than the amountHave?
	// We know that if it's equal we'll just mark the order as filled, return the debited, return the credited, and move on.
	// In what cases would it be greater?
	// Well we don't want to credit them more than they have. So we can only fill it up to a certain amount.
	if amountWantToFill >= a.AmountHave {
		fillRemainder = amountWantToFill - a.AmountHave
		if orderExec, setExec, err = a.GenerateOrderFill(orderID, execPrice); err != nil {
			err = fmt.Errorf("Error generating order fill while generating exec for price: %s", err)
			return
		}
	} else {
		// TODO: Check if this len(orderID) as size is correct
		orderExec = OrderExecution{
			OrderID: make([]byte, len(orderID)),
			Debited: Entry{
				Amount: amountToFill,
				Asset:  debitAsset,
			},
			Credited: Entry{
				Amount: amountWantToFill,
				Asset:  creditAsset,
			},
			NewAmountWant: a.AmountWant - amountToFill,
			NewAmountHave: a.AmountHave - amountWantToFill,
			Filled:        false,
		}
		setExec = SettlementExecution{
			Pubkey: a.Pubkey,
			Debited: Entry{
				Amount: amountToFill,
				Asset:  debitAsset,
			},
			Credited: Entry{
				Amount: amountWantToFill,
				Asset:  creditAsset,
			},
		}
		copy(orderExec.OrderID, orderID)
	}

	// If it's a sell side, price is have/want. So amountToFill * execPrice = amountHave to fill
	// TODO
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
	// nonce [2 bytes]
	// len sig [8 bytes]
	// sig [len sig bytes]
	buf = append(buf, a.Pubkey[:]...)
	buf = append(buf, a.TradingPair.Serialize()...)

	amountHaveBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountHaveBytes, a.AmountHave)
	buf = append(buf, amountHaveBytes[:]...)

	amountWantBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountWantBytes, a.AmountWant)
	buf = append(buf, amountWantBytes[:]...)

	lenSideBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenSideBytes, uint64(len(a.Side)))
	buf = append(buf, lenSideBytes[:]...)

	buf = append(buf, []byte(a.Side)...)
	buf = append(buf, a.AuctionID[:]...)
	buf = append(buf, a.Nonce[:]...)

	lenSigBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenSigBytes, uint64(len(a.Signature)))
	buf = append(buf, lenSigBytes[:]...)

	buf = append(buf, a.Signature[:]...)
	return
}

// SerializeSignable serializes the fields that are hashable, and will be signed. These are also
// what would get verified.
func (a *AuctionOrder) SerializeSignable() (buf []byte) {
	// serializable fields:
	// public key (compressed) [33 bytes]
	// trading pair [2 bytes]
	// amounthave [8 bytes]
	// amountwant [8 bytes]
	// len side [8 bytes]
	// side [len side]
	// auctionID [32 bytes]
	// nonce [2 bytes]
	buf = append(buf, a.Pubkey[:]...)
	buf = append(buf, a.TradingPair.Serialize()...)

	amountHaveBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountHaveBytes, a.AmountHave)
	buf = append(buf, amountHaveBytes[:]...)

	amountWantBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(amountWantBytes, a.AmountWant)
	buf = append(buf, amountWantBytes[:]...)

	lenSideBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenSideBytes, uint64(len(a.Side)))
	buf = append(buf, lenSideBytes[:]...)

	buf = append(buf, []byte(a.Side)...)
	buf = append(buf, a.AuctionID[:]...)
	buf = append(buf, a.Nonce[:]...)
	return
}

// Deserialize deserializes an order into the struct ptr it's being called on
func (a *AuctionOrder) Deserialize(data []byte) (err error) {
	// 33 for pubkey, 26 for the rest, 8 for len side, 4 for min side ("sell" is 4 bytes), 32 for auctionID, 2 for nonce, 8 for siglen
	// bucket is where we put all of the non byte stuff so we can get their length

	// TODO: remove all of this serialization code entirely and use protobufs or something else
	minimumDataLength := len(a.Nonce) +
		len(a.AuctionID) +
		binary.Size(a.AmountWant) +
		binary.Size(a.AmountHave) +
		a.TradingPair.Size() +
		len(a.Pubkey)
	if len(data) < minimumDataLength {
		err = fmt.Errorf("Auction order cannot be less than %d bytes: %s", len(data), err)
		return
	}

	copy(a.Pubkey[:], data[:33])
	data = data[33:]
	var tradingPairBytes [2]byte
	copy(tradingPairBytes[:], data[:2])
	if err = a.TradingPair.Deserialize(tradingPairBytes[:]); err != nil {
		err = fmt.Errorf("Could not deserialize trading pair while deserializing auction order: %s", err)
		return
	}
	data = data[2:]
	a.AmountHave = binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	a.AmountWant = binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	sideLen := binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	a.Side = string(data[:sideLen])
	data = data[sideLen:]
	copy(a.AuctionID[:], data[:32])
	data = data[32:]
	copy(a.Nonce[:], data[:2])
	data = data[2:]
	sigLen := binary.LittleEndian.Uint64(data[:8])
	data = data[8:]
	a.Signature = data[:sigLen]
	data = data[sigLen:]

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

func (a *AuctionOrder) String() string {
	// we ignore error because there's nothing we can do in a String() method
	// to pass on the error other than panic, and I don't want to do that?
	orderMarshalled, _ := json.Marshal(a)
	return string(orderMarshalled)
}
