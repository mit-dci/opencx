package benchclient

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/match"
)

// OrderCommand submits an order synchronously. Uses asynchronous order function
func (cl *BenchClient) OrderCommand(client string, side string, pair string, amountHave uint64, price float64) (reply *cxrpc.SubmitOrderReply, err error) {
	errorChannel := make(chan error, 1)
	replyChannel := make(chan *cxrpc.SubmitOrderReply, 1)
	cl.OrderAsync(client, side, pair, amountHave, price, replyChannel, errorChannel)
	err = <-errorChannel
	if err != nil {
		return
	}
	reply = <-replyChannel
	return
}

// OrderAsync is supposed to be run in a separate goroutine, OrderCommand makes this synchronous however
func (cl *BenchClient) OrderAsync(client string, side string, pair string, amountHave uint64, price float64, replyChan chan *cxrpc.SubmitOrderReply, errChan chan error) {

	errChan <- func() error {
		// TODO: this can be refactored to look more like the rest of the code, it's just using channels and works really well so I don't want to mess with it rn
		orderArgs := new(cxrpc.SubmitOrderArgs)
		orderReply := new(cxrpc.SubmitOrderReply)

		var newOrder match.LimitOrder
		newOrder.Client = client
		newOrder.Side = side

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), cl.PrivKey[:])
		if newOrder.Side != "buy" && newOrder.Side != "sell" {
			return fmt.Errorf("Order's side isn't buy or sell, try again")
		}

		// get the trading pair string from the shell input - third parameter
		err := newOrder.TradingPair.FromString(pair)
		if err != nil {
			return fmt.Errorf("Error getting asset pair from string: \n%s", err)
		}

		newOrder.AmountHave = amountHave

		newOrder.SetAmountWant(price)

		// create e = hash(m)
		sha3 := sha3.New256()
		sha3.Write(newOrder.Serialize())
		e := sha3.Sum(nil)
		// Sign order
		compactSig, err := koblitz.SignCompact(koblitz.S256(), privkey, e, false)

		orderArgs.Signature = compactSig
		orderArgs.Order = &newOrder

		if err = cl.Call("OpencxRPC.SubmitOrder", orderArgs, orderReply); err != nil {
			return fmt.Errorf("Error calling 'SubmitOrder' service method:\n%s", err)
		}

		return nil
	}()

}

// GetPrice calls the getprice rpc command
func (cl *BenchClient) GetPrice(assetString string) (getPriceReply *cxrpc.GetPriceReply, err error) {
	getPriceReply = new(cxrpc.GetPriceReply)
	getPriceArgs := &cxrpc.GetPriceArgs{
		TradingPair: new(match.Pair),
	}

	// get the trading pair string from the shell input - first parameter
	if err = getPriceArgs.TradingPair.FromString(assetString); err != nil {
		return
	}

	if err = cl.Call("OpencxRPC.GetPrice", getPriceArgs, getPriceReply); err != nil {
		return
	}

	return
}

// ViewOrderbook return s the orderbook TODO
func (cl *BenchClient) ViewOrderbook(assetPair string) (viewOrderbookReply *cxrpc.ViewOrderBookReply, err error) {
	viewOrderbookReply = new(cxrpc.ViewOrderBookReply)
	viewOrderBookArgs := &cxrpc.ViewOrderBookArgs{
		TradingPair: new(match.Pair),
	}

	// get the trading pair string from the shell input - first parameter
	if err = viewOrderBookArgs.TradingPair.FromString(assetPair); err != nil {
		return
	}

	// Actually use the RPC Client to call the method
	if err = cl.Call("OpencxRPC.ViewOrderBook", viewOrderBookArgs, viewOrderbookReply); err != nil {
		return
	}

	return
}

// CancelOrder calls the cancel order rpc command
func (cl *BenchClient) CancelOrder(orderID string) (cancelOrderReply *cxrpc.CancelOrderReply, err error) {
	cancelOrderReply = new(cxrpc.CancelOrderReply)
	cancelOrderArgs := &cxrpc.CancelOrderArgs{
		OrderID: orderID,
	}

	// Actually use the RPC Client to call the method
	if err = cl.Call("OpencxRPC.CancelOrder", cancelOrderArgs, cancelOrderReply); err != nil {
		return
	}

	return
}

// GetPairs gets the available trading pairs
func (cl *BenchClient) GetPairs() (getPairsReply *cxrpc.GetPairsReply, err error) {
	getPairsReply = new(cxrpc.GetPairsReply)
	getPairsArgs := new(cxrpc.GetPairsArgs)

	if err = cl.Call("OpencxRPC.GetPairs", getPairsArgs, getPairsReply); err != nil {
		return
	}

	return
}
