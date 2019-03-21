package cxbenchmark

import (
	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// TODO: make this work with public keys. Currently doesn't work.

// PlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func PlaceAndFill(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) {
	for i := 0; i < howMany; i++ {
		// This shouldnt make any change in balance but each account should have at least 2000 satoshis (or the smallest unit in whatever chain)
		bufErrChan := make(chan error, 4)
		orderChan := make(chan *cxrpc.SubmitOrderReply)
		go client.OrderAsync(user1, "buy", pair, 1000, 1.0, orderChan, bufErrChan)
		go client.OrderAsync(user2, "sell", pair, 1000, 1.0, orderChan, bufErrChan)
		go client.OrderAsync(user1, "sell", pair, 2000, 2.0, orderChan, bufErrChan)
		go client.OrderAsync(user2, "buy", pair, 1000, 2.0, orderChan, bufErrChan)

		for i := 0; i < cap(bufErrChan); i++ {
			if err := <-bufErrChan; err != nil {
				logging.Errorf("Error placing and filling: %s", err)
			}
		}
	}
	return
}

// PlaceManyBuy places many orders at once
func PlaceManyBuy(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxrpc.SubmitOrderReply)
	for i := 0; i < howMany; i++ {
		go client.OrderAsync(user1, "buy", pair, 1000, 1.0, orderChan, bufErrChan)
	}
	for i := 0; i < cap(bufErrChan); i++ {
		if err := <-bufErrChan; err != nil {
			logging.Errorf("Error placing many: %s", err)
		}
	}
	return
}

// PlaceManySell places many orders at once
func PlaceManySell(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxrpc.SubmitOrderReply)
	for i := 0; i < howMany; i++ {
		go client.OrderAsync(user1, "sell", pair, 1000, 1.0, orderChan, bufErrChan)
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
	PlaceManyBuy(client, "tester", "othertester", "btc/ltc", varRuns)
	PlaceManyBuy(client, "tester", "othertester", "btc/vtc", varRuns)
	PlaceManyBuy(client, "tester", "othertester", "vtc/ltc", varRuns)
	PlaceManySell(client, "tester", "othertester", "btc/ltc", varRuns)
	PlaceManySell(client, "tester", "othertester", "btc/vtc", varRuns)
	PlaceManySell(client, "tester", "othertester", "vtc/ltc", varRuns)
	return
}

// PlaceFillX runs functions with predone parameters, you only tell it how many times it should run and what client to use
func PlaceFillX(client *benchclient.BenchClient, varRuns int) {
	PlaceAndFill(client, "tester", "othertester", "btc/ltc", varRuns)
	PlaceAndFill(client, "tester", "othertester", "btc/vtc", varRuns)
	PlaceAndFill(client, "tester", "othertester", "vtc/ltc", varRuns)
}
