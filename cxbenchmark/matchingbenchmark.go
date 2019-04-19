package cxbenchmark

import (
	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// TODO: make this work with public keys. Currently doesn't work.

// PlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func PlaceAndFill(client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, pair string, howMany int) {
	for i := 0; i < howMany; i++ {
		// This shouldnt make any change in balance but each account should have at least 2000 satoshis (or the smallest unit in whatever chain)
		bufErrChan := make(chan error, 4)
		orderChan := make(chan *cxrpc.SubmitOrderReply)
		go client1.OrderAsync(client1.PrivKey.PubKey(), "buy", pair, 1000, 1.0, orderChan, bufErrChan)
		go client2.OrderAsync(client2.PrivKey.PubKey(), "sell", pair, 1000, 1.0, orderChan, bufErrChan)
		go client1.OrderAsync(client1.PrivKey.PubKey(), "sell", pair, 2000, 2.0, orderChan, bufErrChan)
		go client2.OrderAsync(client2.PrivKey.PubKey(), "buy", pair, 1000, 2.0, orderChan, bufErrChan)

		for i := 0; i < cap(bufErrChan); i++ {
			if err := <-bufErrChan; err != nil {
				logging.Errorf("Error placing and filling: %s", err)
			}
		}
	}
	return
}

// PlaceManyBuy places many orders at once
func PlaceManyBuy(client *benchclient.BenchClient, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxrpc.SubmitOrderReply)
	for i := 0; i < howMany; i++ {
		go client.OrderAsync(client.PrivKey.PubKey(), "buy", pair, 1000, 1.0, orderChan, bufErrChan)
	}
	for i := 0; i < cap(bufErrChan); i++ {
		if err := <-bufErrChan; err != nil {
			logging.Errorf("Error placing many: %s", err)
		}
	}
	return
}

// PlaceManySell places many orders at once
func PlaceManySell(client *benchclient.BenchClient, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxrpc.SubmitOrderReply)
	for i := 0; i < howMany; i++ {
		go client.OrderAsync(client.PrivKey.PubKey(), "sell", pair, 1000, 1.0, orderChan, bufErrChan)
	}
	for i := 0; i < cap(bufErrChan); i++ {
		if err := <-bufErrChan; err != nil {
			logging.Errorf("Error placing many: %s", err)
		}
	}
	return
}

// PlaceBuySellX runs functions with predone parameters, you only tell it how many times it should run and what client to use
func PlaceBuySellX(client *benchclient.BenchClient, varRuns int) {
	PlaceManyBuy(client, "regtest/litereg", varRuns)
	PlaceManyBuy(client, "regtest/vtcreg", varRuns)
	PlaceManyBuy(client, "vtcreg/litereg", varRuns)
	PlaceManySell(client, "regtest/litereg", varRuns)
	PlaceManySell(client, "regtest/vtcreg", varRuns)
	PlaceManySell(client, "vtcreg/litereg", varRuns)
	return
}

// PlaceFillX runs functions with predone parameters, you only tell it how many times it should run and what client to use
func PlaceFillX(client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, varRuns int) {
	PlaceAndFill(client1, client2, "regtest/litereg", varRuns)
	PlaceAndFill(client1, client2, "regtest/vtcreg", varRuns)
	PlaceAndFill(client1, client2, "vtcreg/litereg", varRuns)
}
