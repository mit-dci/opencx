package cxrpc

import (
	"fmt"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// SubmitOrderArgs holds the args for the submitorder command
type SubmitOrderArgs struct {
	Order *match.LimitOrder
	// Signature is a compact signature so we can do pubkey recovery
	Signature []byte
}

// SubmitOrderReply holds the reply for the submitorder command
type SubmitOrderReply struct {
	OrderID *match.OrderID
}

// SubmitOrder submits an order to the order book or throws an error
func (cl *OpencxRPC) SubmitOrder(args SubmitOrderArgs, reply *SubmitOrderReply) (err error) {

	var orderBytes []byte
	if orderBytes, err = args.Order.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing order for SubmitOrder RPC command: %s", err)
		return
	}

	// hash order.
	sha3 := sha3.New256()
	sha3.Write(orderBytes)
	e := sha3.Sum(nil)

	var sigPubKey *koblitz.PublicKey
	if sigPubKey, _, err = koblitz.RecoverCompact(koblitz.S256(), args.Signature, e); err != nil {
		err = fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
		return
	}

	// try to parse the order pubkey into koblitz
	var orderPubkey *koblitz.PublicKey
	if orderPubkey, err = koblitz.ParsePubKey(args.Order.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Public Key failed parsing check: \n%s", err)
		return
	}

	if !sigPubKey.IsEqual(orderPubkey) {
		err = fmt.Errorf("Pubkey used with signature not equal to the one passed")
		return
	}

	// possible replay attack: if we're using the same pubkey for two exchanges and this is like a feature on the exchange, then an exchange could have you
	// place an order on their exchange, even with a nonce, and then send it over to the other exchange. When you submit an order on one exchange,
	// you essentially submit an order to all of them. But like once we have channels for orders then this isn't a thing anymore because the channel
	// tx's are signed and funding stuff is published on chain
	if reply.OrderID, err = cl.Server.PlaceOrder(args.Order); err != nil {
		err = fmt.Errorf("Error placing order for PlaceOrder RPC command: %s", err)
		return
	}

	var text []byte
	if text, err = reply.OrderID.MarshalText(); err != nil {
		err = fmt.Errorf("Could not marshal text for some reason: %s", err)
		return
	}

	logging.Infof("User %x submitted OrderID %s", sigPubKey.SerializeCompressed(), text)

	return
}

// ViewOrderBookArgs holds the args for the vieworderbook command
type ViewOrderBookArgs struct {
	TradingPair *match.Pair
}

// ViewOrderBookReply holds the reply for the vieworderbook command
type ViewOrderBookReply struct {
	Orderbook map[float64][]*match.LimitOrderIDPair
}

// ViewOrderBook handles the vieworderbook command
func (cl *OpencxRPC) ViewOrderBook(args ViewOrderBookArgs, reply *ViewOrderBookReply) (err error) {

	if reply.Orderbook, err = cl.Server.ViewOrderbook(args.TradingPair); err != nil {
		err = fmt.Errorf("Error with server ViewOrderbook for ViewOrderbook RPC command: %s", err)
		return
	}

	return
}

// GetPriceArgs holds the args for the GetPrice command
type GetPriceArgs struct {
	TradingPair *match.Pair
}

// GetPriceReply holds the reply for the GetPrice command
type GetPriceReply struct {
	Price float64
}

// GetPrice returns the price for the specified asset
func (cl *OpencxRPC) GetPrice(args GetPriceArgs, reply *GetPriceReply) (err error) {

	if reply.Price, err = cl.Server.GetPrice(args.TradingPair); err != nil {
		err = fmt.Errorf("Error calculating price for GetPrice RPC command: %s", err)
		return
	}

	return
}

// TODO: any man in the middle of a GetOrder communication can replay the same thing as a Cancel communication.

// CancelOrderArgs holds the args for the CancelOrder command
type CancelOrderArgs struct {
	OrderID   string
	Signature []byte
}

// CancelOrderReply holds the args for the CancelOrder command
type CancelOrderReply struct {
	// empty
}

// CancelOrder cancels the order
func (cl *OpencxRPC) CancelOrder(args CancelOrderArgs, reply *CancelOrderReply) (err error) {

	// hash order.
	sha3 := sha3.New256()
	sha3.Write([]byte(args.OrderID))
	e := sha3.Sum(nil)

	logging.Infof("Checking cancel signature")
	var sigPubKey *koblitz.PublicKey
	if sigPubKey, _, err = koblitz.RecoverCompact(koblitz.S256(), args.Signature, e); err != nil {
		err = fmt.Errorf("Error verifying cancel, invalid signature: \n%s", err)
		return
	}

	var unmarshalledOrderID *match.OrderID = new(match.OrderID)
	if err = unmarshalledOrderID.UnmarshalText([]byte(args.OrderID)); err != nil {
		err = fmt.Errorf("Error unmarshalling text for Order ID in CancelOrder RPC: %s", err)
		return
	}

	var orderPair *match.LimitOrderIDPair
	if orderPair, err = cl.Server.GetOrder(unmarshalledOrderID); err != nil {
		err = fmt.Errorf("Error calling GetOrder in CancelOrder RPC: %s", err)
		return
	}

	// try to parse the order pubkey into koblitz
	var orderPubKey *koblitz.PublicKey
	if orderPubKey, err = koblitz.ParsePubKey(orderPair.Order.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Public Key failed parsing check: \n%s", err)
		return
	}

	if !sigPubKey.IsEqual(orderPubKey) {
		err = fmt.Errorf("Pubkey used with signature not equal to the one passed")
		return
	}

	// More defensive programming
	if *orderPair.OrderID != *unmarshalledOrderID {
		err = fmt.Errorf("Error: OrderID from returned order differs from the order ID passed as argument")
		return
	}

	if err = cl.Server.CancelOrder(orderPair); err != nil {
		err = fmt.Errorf("Error cancelling order for CancelOrder RPC command: %s", err)
		return
	}

	return
}

// GetPairsArgs holds the args for the GetPairs command
type GetPairsArgs struct {
	// empty
}

// GetPairsReply holds the reply for the GetPairs command
type GetPairsReply struct {
	PairList []string
}

// GetPairs gets all the pairs as nice strings
func (cl *OpencxRPC) GetPairs(args GetPairsArgs, reply *GetPairsReply) (err error) {
	// just go through all the pairs and prettily print them
	for _, pair := range cl.Server.GetPairs() {
		reply.PairList = append(reply.PairList, pair.PrettyString())
	}

	return
}

// GetOrderArgs holds the args for the GetOrder command
type GetOrderArgs struct {
	OrderID   string
	Signature []byte
}

// GetOrderReply holds the reply for the GetOrder command
type GetOrderReply struct {
	Order *match.LimitOrderIDPair
}

// GetOrder gets an order based on orderID
func (cl *OpencxRPC) GetOrder(args GetOrderArgs, reply *GetOrderReply) (err error) {
	// hash order id.
	sha3 := sha3.New256()
	sha3.Write([]byte(args.OrderID))
	e := sha3.Sum(nil)

	logging.Infof("Checking getorder signature")
	var sigPubKey *koblitz.PublicKey
	if sigPubKey, _, err = koblitz.RecoverCompact(koblitz.S256(), args.Signature, e); err != nil {
		err = fmt.Errorf("Error verifying getorder, invalid signature: \n%s", err)
		return
	}

	var unmarshalledOrderID *match.OrderID = new(match.OrderID)
	if err = unmarshalledOrderID.UnmarshalText([]byte(args.OrderID)); err != nil {
		err = fmt.Errorf("Could not unmarshal order ID for GetOrder command: %s", err)
		return
	}

	if reply.Order, err = cl.Server.GetOrder(unmarshalledOrderID); err != nil {
		err = fmt.Errorf("Error getting order from server for GetOrder RPC command: %s", err)
		return
	}

	// try to parse the order pubkey into koblitz
	var orderPubKey *koblitz.PublicKey
	if orderPubKey, err = koblitz.ParsePubKey(reply.Order.Order.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Public Key failed parsing check for GetOrder RPC command: %s", err)
		return
	}

	if !sigPubKey.IsEqual(orderPubKey) {
		err = fmt.Errorf("Pubkey used with signature not equal to the one passed")
		return
	}

	return
}

// GetOrdersForPubkeyArgs holds the args for the GetOrdersForPubkey command
type GetOrdersForPubkeyArgs struct {
	Signature []byte
}

// GetOrdersForPubkeyReply holds the reply for the GetOrdersForPubkey command
type GetOrdersForPubkeyReply struct {
	Orders []*match.LimitOrderIDPair
}

// GetOrdersForPubkey gets the orders for the pubkey which has signed the getOrdersString
func (cl *OpencxRPC) GetOrdersForPubkey(args GetOrdersForPubkeyArgs, reply *GetOrdersForPubkeyReply) (err error) {
	var pubkey *koblitz.PublicKey
	if pubkey, err = cl.Server.GetOrdersStringVerify(args.Signature); err != nil {
		return
	}

	if reply.Orders, err = cl.Server.GetOrdersForPubkey(pubkey); err != nil {
		return
	}

	return
}
