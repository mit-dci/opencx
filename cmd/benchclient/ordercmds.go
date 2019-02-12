package benchclient

import (
	"bytes"
	"fmt"

	"github.com/mit-dci/opencx/logging"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/match"

	"github.com/olekukonko/tablewriter"
)

// OrderCommand submits an order (for now) TODO
func (cl *BenchClient) OrderCommand(client string, side string, pair string, amountHave uint64, price float64) error {
	orderArgs := new(cxrpc.SubmitOrderArgs)
	orderReply := new(cxrpc.SubmitOrderReply)

	var newOrder match.LimitOrder
	newOrder.Client = client
	newOrder.Side = side
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

	orderArgs.Order = &newOrder
	err = cl.Call("OpencxRPC.SubmitOrder", orderArgs, orderReply)
	if err != nil {
		return fmt.Errorf("Error calling 'SubmitOrder' service method:\n%s", err)
	}

	logging.Infof("Order submitted successfully\n")
	return nil
}

// GetPrice calls the getprice rpc command
func (cl *BenchClient) GetPrice(assetString string) error {
	var err error

	getPriceArgs := new(cxrpc.GetPriceArgs)
	getPriceReply := new(cxrpc.GetPriceReply)

	// can't be a nil pointer to call methods on it
	getPriceArgs.TradingPair = new(match.Pair)

	// get the trading pair string from the shell input - first parameter
	if err = getPriceArgs.TradingPair.FromString(assetString); err != nil {
		return err
	}

	if err = cl.Call("OpencxRPC.GetPrice", getPriceArgs, getPriceReply); err != nil {
		return err
	}

	logging.Infof("Price: %f\n", getPriceReply.Price)
	return nil
}

// ViewOrderbook return s the orderbook TODO
func (cl *BenchClient) ViewOrderbook(assetPair string) error {
	var err error

	viewOrderBookArgs := new(cxrpc.ViewOrderBookArgs)
	viewOrderBookReply := new(cxrpc.ViewOrderBookReply)

	// can't be a nil pointer to call methods on it
	viewOrderBookArgs.TradingPair = new(match.Pair)

	// get the trading pair string from the shell input - first parameter
	err = viewOrderBookArgs.TradingPair.FromString(assetPair)
	if err != nil {
		return err
	}

	// Actually use the RPC Client to call the method
	err = cl.Call("OpencxRPC.ViewOrderBook", viewOrderBookArgs, viewOrderBookReply)
	if err != nil {
		return fmt.Errorf("Error calling 'ViewOrderBook' service method:\n%s", err)
	}

	// Build the table
	var data [][]string
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"orderID", "price", "volume", "side"})

	// get all buy orders and add to table
	for _, buyOrder := range viewOrderBookReply.BuyOrderBook {
		buyPrice, err := buyOrder.Price()
		if err != nil {
			return err
		}

		// convert stuff to strings
		strPrice := fmt.Sprintf("%f", buyPrice)
		strVolume := fmt.Sprintf("%d", buyOrder.AmountHave)
		// append to the table
		data = append(data, []string{buyOrder.OrderID, strPrice, strVolume, buyOrder.Side})
	}

	// get all the sell orders and add to table
	for _, sellOrder := range viewOrderBookReply.SellOrderBook {
		sellPrice, err := sellOrder.Price()
		if err != nil {
			return err
		}

		// convert stuff to strings
		strPrice := fmt.Sprintf("%f", sellPrice)
		strVolume := fmt.Sprintf("%d", sellOrder.AmountHave)
		// append to the table
		data = append(data, []string{sellOrder.OrderID, strPrice, strVolume, sellOrder.Side})
	}

	// render the table
	table.AppendBulk(data)
	table.Render()

	// actually print out table stored in buffer
	logging.Infof("\n%s\n", buf.String())
	return nil
}
