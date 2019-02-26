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

	pubkey, _, err := koblitz.RecoverCompact(koblitz.S256(), args.Signature, e)
	if err != nil {
		return fmt.Errorf("Error verifying order, invalid signature: \n%s", err)
	}

	logging.Infof("Pubkey %x submitted order", pubkey.SerializeUncompressed())

	cl.Server.LockOrders()
	cl.Server.OrderMap[args.Order.TradingPair] = append(cl.Server.OrderMap[args.Order.TradingPair], args.Order)
	cl.Server.UnlockOrders()

	cl.Server.LockIngests()
	var id string
	if id, err = cl.Server.OpencxDB.PlaceOrder(args.Order); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		// logging.Errorf("Error with matching: \n%s", err)
		cl.Server.UnlockIngests()
		err = fmt.Errorf("Error placing order while submitting order: \n%s", err)
		return
	}
	cl.Server.UnlockIngests()

	reply.OrderID = id
	orderPrice, err := args.Order.Price()
	if err != nil {
		return fmt.Errorf("Error submitting order and calculating price: \n%s", err)
	}

	startedChan := make(chan bool, 1)
	go cl.Server.MatchingRoutine(startedChan, &args.Order.TradingPair, orderPrice)

	return
}

// ViewOrderBookArgs holds the args for the vieworderbook command
type ViewOrderBookArgs struct {
	TradingPair *match.Pair
}

// ViewOrderBookReply holds the reply for the vieworderbook command
type ViewOrderBookReply struct {
	SellOrderBook []*match.LimitOrder
	BuyOrderBook  []*match.LimitOrder
}

// ViewOrderBook handles the vieworderbook command
func (cl *OpencxRPC) ViewOrderBook(args ViewOrderBookArgs, reply *ViewOrderBookReply) (err error) {

	cl.Server.LockIngests()
	if reply.SellOrderBook, reply.BuyOrderBook, err = cl.Server.OpencxDB.ViewOrderBook(args.TradingPair); err != nil {
		return err
	}
	cl.Server.UnlockIngests()

	return nil
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

// CancelOrderArgs holds the args for the CancelOrder command
type CancelOrderArgs struct {
	OrderID string
}

// CancelOrderReply holds the args for the CancelOrder command
type CancelOrderReply struct {
	// empty
}

// CancelOrder cancels the order
func (cl *OpencxRPC) CancelOrder(args CancelOrderArgs, reply *CancelOrderReply) (err error) {
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
	for _, pair := range cl.Server.PairsArray {
		reply.PairList = append(reply.PairList, pair.PrettyString())
	}

	return
}
