package cxbenchmark

import (
	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// PlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func PlaceAndFill(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) (numRuns int) {
	for i := 0; i < howMany; i++ {
		// This shouldnt make any change in balance but each account should have at least 2000 satoshis (or the smallest unit in whatever chain)
		bufErrChan := make(chan error, 4)
		go client.OrderAsync(user1, "buy", pair, 1000, 1.0, bufErrChan)
		go client.OrderAsync(user2, "sell", pair, 1000, 1.0, bufErrChan)
		go client.OrderAsync(user1, "sell", pair, 2000, 2.0, bufErrChan)
		go client.OrderAsync(user2, "buy", pair, 1000, 2.0, bufErrChan)

		for i := 0; i < cap(bufErrChan); i++ {
			if err := <-bufErrChan; err != nil {
				logging.Errorf("Error placing and filling: %s", err)
			}
			numRuns++
		}

	}
	return
}
