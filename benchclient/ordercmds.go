package benchclient

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// OrderCommand submits an order synchronously. Uses asynchronous order function
func (cl *BenchClient) OrderCommand(pubkey *koblitz.PublicKey, side match.Side, pair string, amountHave uint64, price float64) (reply *cxrpc.SubmitOrderReply, err error) {
	errorChannel := make(chan error, 1)
	replyChannel := make(chan *cxrpc.SubmitOrderReply, 1)
	go cl.OrderAsync(pubkey, side, pair, amountHave, price, replyChannel, errorChannel)
	// wait on either the reply or error, whichever comes first. If error is nil wait for reply. That's why the for loop is there. We don't care if the reply is nil, it shouldn't be, but that's sort of just so go-vet doesn't yell at us for having an unreachable return.
	for reply == nil {
		select {
		case reply = <-replyChannel:
			return
		case err = <-errorChannel:
			if err != nil {
				return
			}
		}
	}

	return
}

// OrderAsync is supposed to be run in a separate goroutine, OrderCommand makes this synchronous however
func (cl *BenchClient) OrderAsync(pubkey *koblitz.PublicKey, side match.Side, pair string, amountHave uint64, price float64, replyChan chan *cxrpc.SubmitOrderReply, errChan chan error) {

	if cl.PrivKey == nil {
		errChan <- fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	errChan <- func() (err error) {
		// TODO: this can be refactored to look more like the rest of the code, it's just using channels and works really well so I don't want to mess with it rn
		orderArgs := new(cxrpc.SubmitOrderArgs)
		orderReply := new(cxrpc.SubmitOrderReply)

		var newOrder match.LimitOrder
		copy(newOrder.Pubkey[:], pubkey.SerializeCompressed())
		newOrder.Side = side

		// get the trading pair string from the shell input - third parameter
		if err = newOrder.TradingPair.FromString(pair); err != nil {
			err = fmt.Errorf("Error getting asset pair from string: \n%s", err)
			return
		}

		newOrder.AmountHave = amountHave
		newOrder.AmountWant = uint64(price * float64(amountHave))

		var newOrderBytes []byte
		if newOrderBytes, err = newOrder.Serialize(); err != nil {
			err = fmt.Errorf("Error serializing new order: %s", err)
			return
		}

		// create e = hash(m)
		sha3 := sha3.New256()
		sha3.Write(newOrderBytes)
		e := sha3.Sum(nil)

		// Sign order
		var compactSig []byte
		if compactSig, err = koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false); err != nil {
			return
		}

		orderArgs.Signature = compactSig
		orderArgs.Order = &newOrder

		if err = cl.Call("OpencxRPC.SubmitOrder", orderArgs, orderReply); err != nil {
			err = fmt.Errorf("Error calling 'SubmitOrder' service method:\n%s", err)
			return
		}

		replyChan <- orderReply

		return
	}()

	return
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

// ViewOrderbook returns the orderbook
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

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	cancelOrderReply = new(cxrpc.CancelOrderReply)
	cancelOrderArgs := &cxrpc.CancelOrderArgs{
		OrderID: orderID,
	}

	// create e = hash(m)
	sha3 := sha3.New256()
	sha3.Write([]byte(cancelOrderArgs.OrderID))
	e := sha3.Sum(nil)

	// Sign order
	compactSig, err := koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false)

	cancelOrderArgs.Signature = compactSig

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

// AuctionOrderCommand submits an order synchronously. Uses asynchronous order function
func (cl *BenchClient) AuctionOrderCommand(pubkey *koblitz.PublicKey, side string, pair string, amountHave uint64, price float64, t uint64, auctionID [32]byte) (reply *cxauctionrpc.SubmitPuzzledOrderReply, err error) {
	errorChannel := make(chan error, 1)
	replyChannel := make(chan *cxauctionrpc.SubmitPuzzledOrderReply, 1)
	go cl.AuctionOrderAsync(pubkey, side, pair, amountHave, price, t, auctionID, replyChannel, errorChannel)
	// wait on either the reply or error, whichever comes first. If error is nil wait for reply. That's why the for loop is there. We don't care if the reply is nil, it shouldn't be, but that's sort of just so go-vet doesn't yell at us for having an unreachable return.
	for reply == nil {
		select {
		case reply = <-replyChannel:
			return
		case err = <-errorChannel:
			if err != nil {
				return
			}
		}
	}

	return
}

// AuctionOrderAsync is supposed to be run in a separate goroutine, AuctionOrderCommand makes this synchronous however
func (cl *BenchClient) AuctionOrderAsync(pubkey *koblitz.PublicKey, side string, pair string, amountHave uint64, price float64, t uint64, auctionID [32]byte, replyChan chan *cxauctionrpc.SubmitPuzzledOrderReply, errChan chan error) {

	if cl.PrivKey == nil {
		errChan <- fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	errChan <- func() (err error) {
		// TODO: this can be refactored to look more like the rest of the code, it's just using channels and works really well so I don't want to mess with it rn
		orderArgs := new(cxauctionrpc.SubmitPuzzledOrderArgs)
		orderReply := new(cxauctionrpc.SubmitPuzzledOrderReply)

		var newAuctionOrder match.AuctionOrder
		copy(newAuctionOrder.Pubkey[:], pubkey.SerializeCompressed())
		newAuctionOrder.Side = side

		// check that the sides are correct
		if newAuctionOrder.Side != "buy" && newAuctionOrder.Side != "sell" {
			err = fmt.Errorf("AuctionOrder's side isn't buy or sell, try again")
			return
		}

		// get the trading pair string from the shell input - third parameter
		if err = newAuctionOrder.TradingPair.FromString(pair); err != nil {
			err = fmt.Errorf("Error getting asset pair from string: \n%s", err)
			return
		}

		newAuctionOrder.AmountHave = amountHave
		newAuctionOrder.AuctionID = auctionID

		newAuctionOrder.SetAmountWant(price)

		// create e = hash(m)
		sha3 := sha3.New256()
		sha3.Write(newAuctionOrder.SerializeSignable())
		e := sha3.Sum(nil)

		// Sign order
		var compactSig []byte
		if compactSig, err = koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false); err != nil {
			return
		}

		logging.Infof("Order time: %d", t)

		newAuctionOrder.Signature = compactSig
		var order *match.EncryptedAuctionOrder
		if order, err = newAuctionOrder.TurnIntoEncryptedOrder(t); err != nil {
			err = fmt.Errorf("Error turning order into puzzle before submitting: %s", err)
			return
		}

		if orderArgs.EncryptedOrderBytes, err = order.Serialize(); err != nil {
			err = fmt.Errorf("Error when trying to serialize the order: %s", err)
			return
		}

		if err = cl.Call("OpencxAuctionRPC.SubmitPuzzledOrder", orderArgs, orderReply); err != nil {
			err = fmt.Errorf("Error calling 'SubmitPuzzledOrder' service method:\n%s", err)
			return
		}

		replyChan <- orderReply

		return
	}()

	return
}
