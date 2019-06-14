package cxrpc

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
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
	OrderID string
}

// SubmitOrder submits an order to the order book or throws an error
func (cl *OpencxRPC) SubmitOrder(args SubmitOrderArgs, reply *SubmitOrderReply) (err error) {

	// hash order.
	sha3 := sha3.New256()
	sha3.Write(args.Order.Serialize())
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
	cl.Server.LockIngests()
	if reply.OrderID, err = cl.Server.OpencxDB.PlaceOrder(args.Order); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error placing order while submitting order: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()

	logging.Infof("User %x submitted OrderID %s", sigPubKey.SerializeCompressed(), reply.OrderID)

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

	cl.Server.LockIngests()
	if reply.Orderbook, err = cl.Server.OpencxDB.ViewLimitOrderBook(args.TradingPair); err != nil {
		cl.Server.UnlockIngests()
		return
	}
	cl.Server.UnlockIngests()

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
	cl.Server.LockIngests()
	// reply.Price = cl.Server.OpencxDB.GetPrice(args.TradingPair.String())

	if reply.Price, err = cl.Server.OpencxDB.CalculatePrice(args.TradingPair); err != nil {
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error calculating price: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()
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

	cl.Server.LockIngests()
	var order *match.LimitOrder
	if order, err = cl.Server.OpencxDB.GetOrder(args.OrderID); err != nil {
		return
	}
	cl.Server.LockIngests()

	// try to parse the order pubkey into koblitz
	var orderPubKey *koblitz.PublicKey
	if orderPubKey, err = koblitz.ParsePubKey(order.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Public Key failed parsing check: \n%s", err)
		return
	}

	if !sigPubKey.IsEqual(orderPubKey) {
		err = fmt.Errorf("Pubkey used with signature not equal to the one passed")
		return
	}

	cl.Server.LockIngests()
	if err = cl.Server.OpencxDB.CancelOrder(args.OrderID); err != nil {
		cl.Server.UnlockIngests()
		return
	}
	cl.Server.UnlockIngests()

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

// GetPairs gets all the pairs
func (cl *OpencxRPC) GetPairs(args GetPairsArgs, reply *GetPairsReply) (err error) {
	// just go through all the pairs and prettily print them
	for _, pair := range cl.Server.OpencxDB.GetPairs() {
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
	Order *match.LimitOrder
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

	cl.Server.LockIngests()
	if reply.Order, err = cl.Server.OpencxDB.GetOrder(args.OrderID); err != nil {
		return
	}
	cl.Server.UnlockIngests()

	// try to parse the order pubkey into koblitz
	var orderPubKey *koblitz.PublicKey
	if orderPubKey, err = koblitz.ParsePubKey(reply.Order.Pubkey[:], koblitz.S256()); err != nil {
		err = fmt.Errorf("Public Key failed parsing check: \n%s", err)
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
	Orders map[float64][]*match.LimitOrderIDPair
}

// GetOrdersForPubkey gets the orders for the pubkey which has signed the getOrdersString
func (cl *OpencxRPC) GetOrdersForPubkey(args GetOrdersForPubkeyArgs, reply *GetOrdersForPubkeyReply) (err error) {
	var pubkey *koblitz.PublicKey
	if pubkey, err = cl.Server.GetOrdersStringVerify(args.Signature); err != nil {
		return
	}

	cl.Server.LockIngests()
	if reply.Orders, err = cl.Server.OpencxDB.GetOrdersForPubkey(pubkey); err != nil {
		return
	}
	cl.Server.UnlockIngests()

	return
}
