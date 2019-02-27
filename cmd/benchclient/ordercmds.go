package benchclient

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
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
		err = cl.Call("OpencxRPC.SubmitOrder", orderArgs, orderReply)
		if err != nil {
			return fmt.Errorf("Error calling 'SubmitOrder' service method:\n%s", err)
		}

		// logging.Infof("Order submitted successfully\n")
		return nil
	}()

}

// GetPrice calls the getprice rpc command
func (cl *BenchClient) GetPrice(assetString string) (reply *cxrpc.GetPriceReply, err error) {

	getPriceArgs := new(cxrpc.GetPriceArgs)

	// can't be a nil pointer to call methods on it
	getPriceArgs.TradingPair = new(match.Pair)

	// get the trading pair string from the shell input - first parameter
	if err = getPriceArgs.TradingPair.FromString(assetString); err != nil {
		return
	}

	if err = cl.Call("OpencxRPC.GetPrice", getPriceArgs, reply); err != nil {
		return
	}

	return
}

// ViewOrderbook return s the orderbook TODO
func (cl *BenchClient) ViewOrderbook(assetPair string) (reply *cxrpc.ViewOrderBookReply, err error) {
	viewOrderBookArgs := new(cxrpc.ViewOrderBookArgs)

	// can't be a nil pointer to call methods on it
	viewOrderBookArgs.TradingPair = new(match.Pair)

	// get the trading pair string from the shell input - first parameter

	if err = viewOrderBookArgs.TradingPair.FromString(assetPair); err != nil {
		return nil, err
	}

	// Actually use the RPC Client to call the method

	if err = cl.Call("OpencxRPC.ViewOrderBook", viewOrderBookArgs, reply); err != nil {
		err = fmt.Errorf("Error calling 'ViewOrderBook' service method:\n%s", err)
		return
	}

	return
}

// CancelOrder calls the cancel order rpc command
func (cl *BenchClient) CancelOrder(orderID string) (err error) {
	cancelOrderArgs := &cxrpc.CancelOrderArgs{
		OrderID: orderID,
	}
	cancelOrderReply := new(cxrpc.CancelOrderReply)

	// Actually use the RPC Client to call the method
	if err = cl.Call("OpencxRPC.CancelOrder", cancelOrderArgs, cancelOrderReply); err != nil {
		err = fmt.Errorf("Error calling 'CancelOrder' service method:\n%s", err)
		return
	}

	logging.Infof("Cancelled order successfully")

	return
}

// GetPairs gets the available trading pairs
func (cl *BenchClient) GetPairs() (reply *cxrpc.GetPairsReply, err error) {
	getPairsArgs := new(cxrpc.GetPairsArgs)

	if err = cl.Call("OpencxRPC.GetPairs", getPairsArgs, reply); err != nil {
		err = fmt.Errorf("Error calling 'GetPairs' service method:\n%s", err)
		return
	}

	return
}
