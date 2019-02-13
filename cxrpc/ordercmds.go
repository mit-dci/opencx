package cxrpc

import (
	"github.com/mit-dci/opencx/match"
)

// SubmitOrderArgs holds the args for the submitorder command
type SubmitOrderArgs struct {
	Order *match.LimitOrder
}

// SubmitOrderReply holds the reply for the submitorder command
type SubmitOrderReply struct {
	// TODO empty for now
}

// SubmitOrder submits an order to the order book or throws an error
func (cl *OpencxRPC) SubmitOrder(args SubmitOrderArgs, reply *SubmitOrderReply) error {

	cl.Server.LockOrders()
	cl.Server.OrderMap[args.Order.TradingPair] = append(cl.Server.OrderMap[args.Order.TradingPair], args.Order)
	cl.Server.UnlockOrders()

	cl.Server.LockIngests()
	if err := cl.Server.OpencxDB.PlaceOrder(args.Order); err != nil {
		// gotta put these here cause if it errors out then oops just locked everything
		// logging.Errorf("Error with matching: \n%s", err)
		cl.Server.UnlockIngests()
		return err
	}
	cl.Server.UnlockIngests()
	return nil
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
func (cl *OpencxRPC) ViewOrderBook(args ViewOrderBookArgs, reply *ViewOrderBookReply) error {
	var err error
	if reply.SellOrderBook, reply.BuyOrderBook, err = cl.Server.OpencxDB.ViewOrderBook(args.TradingPair); err != nil {
		return err
	}

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
func (cl *OpencxRPC) GetPrice(args GetPriceArgs, reply *GetPriceReply) error {
	reply.Price = cl.Server.OpencxDB.GetPrice(args.TradingPair.String())
	return nil
}
