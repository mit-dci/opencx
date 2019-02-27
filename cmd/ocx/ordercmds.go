package main

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/mit-dci/opencx/logging"

	"github.com/mit-dci/opencx/cxrpc"

	"github.com/olekukonko/tablewriter"
)

// OrderCommand submits an order (for now) TODO
func (cl *openCxClient) OrderCommand(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	client := args[0]
	side := args[1]
	pair := args[2]
	amountHave, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return fmt.Errorf("Error parsing amountHave, please enter something valid:\n%s", err)
	}
	price, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return fmt.Errorf("Error parsing price: \n%s", err)
	}
	var reply *cxrpc.SubmitOrderReply
	if reply, err = cl.RPCClient.OrderCommand(client, side, pair, amountHave, price); err != nil {
		return
	}

	logging.Infof("Submitted order successfully, orderID: %s", reply.OrderID)
	return nil
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

// CancelOrder calls the cancel order rpc command
func (cl *openCxClient) CancelOrder(args []string) (err error) {
	orderID := args[0]
	if err = cl.RPCClient.CancelOrder(orderID); err != nil {
		return
	}

	logging.Infof("Cancelled order successfully")
	return
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
