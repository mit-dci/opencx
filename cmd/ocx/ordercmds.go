package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/logging"

	"github.com/mit-dci/opencx/cxrpc"

	"github.com/olekukonko/tablewriter"
)

var placeOrderCommand = &Command{
	Format: fmt.Sprintf("%s%s%s%s%s\n", lnutil.Red("placeorder"), lnutil.ReqColor("side"), lnutil.ReqColor("pair"), lnutil.ReqColor("amounthave"), lnutil.ReqColor("price")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Submit a order with side \"buy\" or side \"sell\", for pair \"asset1\"/\"asset2\", where you give up amounthave of \"asset1\" (if on buy side) or \"asset2\" if on sell side, for the other token at a specific price.",
		"This will return an order ID which can be used as input to cancelorder, or getorder.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Place an order on the exchange."),
}

// OrderCommand submits an order (for now)
func (cl *openCxClient) OrderCommand(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	side := args[0]
	pair := args[1]

	amountHave, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing amountHave, please enter something valid:\n%s", err)
	}

	price, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return fmt.Errorf("Error parsing price: \n%s", err)
	}

	var pubkey *koblitz.PublicKey
	if pubkey, err = cl.RetreivePublicKey(); err != nil {
		return
	}

	var reply *cxrpc.SubmitOrderReply
	if reply, err = cl.RPCClient.OrderCommand(pubkey, side, pair, amountHave, price); err != nil {
		return
	}

	logging.Infof("Submitted order successfully, orderID: %s", reply.OrderID)
	return nil
}

var getPriceCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.Red("getprice"), lnutil.ReqColor("pair")),
	Description: fmt.Sprintf("%s\n",
		"Get the price of the input asset pair.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get the price of the input asset pair."),
}

// GetPrice prints the price for the asset
func (cl *openCxClient) GetPrice(args []string) (err error) {
	assetString := args[0]

	var getPriceReply *cxrpc.GetPriceReply
	if getPriceReply, err = cl.RPCClient.GetPrice(assetString); err != nil {
		return
	}

	logging.Infof("Price: %f %s\n", getPriceReply.Price, assetString)
	return nil
}

var viewOrderbookCommand = &Command{
	Format: fmt.Sprintf("%s%s%s\n", lnutil.Red("vieworderbook"), lnutil.ReqColor("pair"), lnutil.OptColor("side")),
	Description: fmt.Sprintf("%s\n",
		"View orderbook for pair, with optional side.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "View orderbook for pair, with optional side."),
}

// ViewOrderbook prints the orderbook
func (cl *openCxClient) ViewOrderbook(args []string) (err error) {

	pair := args[0]
	var viewOrderbookReply *cxrpc.ViewOrderBookReply
	if viewOrderbookReply, err = cl.RPCClient.ViewOrderbook(pair); err != nil {
		return
	}

	// Build the table
	var data [][]string
	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"orderID", "price", "volume", "side"})

	// get all buy orders and add to table
	for _, buyOrder := range viewOrderbookReply.BuyOrderBook {
		var buyPrice float64
		if buyPrice, err = buyOrder.Price(); err != nil {
			return
		}

		// convert stuff to strings
		strPrice := fmt.Sprintf("%f", buyPrice)
		strVolume := fmt.Sprintf("%d", buyOrder.AmountHave)
		// append to the table
		data = append(data, []string{buyOrder.OrderID, strPrice, strVolume, buyOrder.Side})
	}

	// get all the sell orders and add to table
	for _, sellOrder := range viewOrderbookReply.SellOrderBook {
		var sellPrice float64
		if sellPrice, err = sellOrder.Price(); err != nil {
			return
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
	return
}

var cancelOrderCommand = &Command{
	Format: fmt.Sprintf("%s%s\n", lnutil.Red("cancelorder"), lnutil.ReqColor("orderID")),
	Description: fmt.Sprintf("%s\n",
		"Cancel order with orderID.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Cancel order with orderID."),
}

// CancelOrder calls the cancel order rpc command
func (cl *openCxClient) CancelOrder(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}
	orderID := args[0]

	// remove this and _ when cancel order has returns
	// var cancelOrderReply *cxrpc.CancelOrderReply
	if _, err = cl.RPCClient.CancelOrder(orderID); err != nil {
		return
	}

	logging.Infof("Cancelled order successfully")
	return
}

var getPairsCommand = &Command{
	Format: fmt.Sprintf("%s\n", lnutil.Red("getpairs")),
	Description: fmt.Sprintf("%s\n",
		"Get all available trading pairs.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Get all available trading pairs."),
}

// GetPairs gets the available trading pairs
func (cl *openCxClient) GetPairs() (err error) {
	var getPairsReply *cxrpc.GetPairsReply
	if getPairsReply, err = cl.RPCClient.GetPairs(); err != nil {
		return
	}

	logging.Infof("List of valid trading pairs: ")
	for _, pair := range getPairsReply.PairList {
		logging.Infof("%s", pair)
	}

	return
}
