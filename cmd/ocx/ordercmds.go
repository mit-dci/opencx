package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/Rjected/lit/lnutil"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"

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
func (cl *ocxClient) OrderCommand(args []string) (err error) {
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
	if pubkey, err = cl.RetrievePublicKey(); err != nil {
		return
	}

	var orderSide *match.Side = new(match.Side)
	if err = orderSide.FromString(side); err != nil {
		err = fmt.Errorf("Error getting side from string for OrderCommand: %s", err)
		return
	}

	var reply *cxrpc.SubmitOrderReply
	if reply, err = cl.RPCClient.OrderCommand(pubkey, *orderSide, pair, amountHave, price); err != nil {
		return
	}

	var text []byte
	if text, err = reply.OrderID.MarshalText(); err != nil {
		err = fmt.Errorf("Could not marshal to text for some reason: %s", err)
		return
	}

	logging.Infof("Submitted order successfully, orderID: %s", text)
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
func (cl *ocxClient) GetPrice(args []string) (err error) {
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
func (cl *ocxClient) ViewOrderbook(args []string) (err error) {

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
	for pr, orderList := range viewOrderbookReply.Orderbook {
		for _, order := range orderList {

			if order.Price != pr {
				warnDiscrepancy := `
					WARNING: Price returned by exchange in map does not equal
					         the price that is recognized in the order. This
							 should be the same, and there may be some foul
							 play done by the exchange.
				`
				logging.Errorf(warnDiscrepancy)
			}

			// convert stuff to strings
			strOrderID := fmt.Sprintf("%x", order.OrderID)
			strPrice := fmt.Sprintf("%f", order.Price)
			strVolume := fmt.Sprintf("%d", order.Order.AmountHave)
			// append to the table
			data = append(data, []string{strOrderID, strPrice, strVolume, order.Order.Side.String()})
		}
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
func (cl *ocxClient) CancelOrder(args []string) (err error) {
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
func (cl *ocxClient) GetPairs() (err error) {
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
