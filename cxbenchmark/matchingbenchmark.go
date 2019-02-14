package cxbenchmark

import (
	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// PlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func PlaceAndFill(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) (numRuns int) {
	for pr := 0.1; pr < 1; pr = pr + 0.1 {
		// This shouldnt make any change in balance but each account should have at least 500 satoshis (of the smallest unit in whatever chain)
		bufErrChan := make(chan error, 4)
		go client.OrderAsync(user1, "buy", pair, 10, pr, bufErrChan)
		go client.OrderAsync(user2, "sell", pair, 10, pr, bufErrChan)
		go client.OrderAsync(user1, "sell", pair, 10, pr, bufErrChan)
		go client.OrderAsync(user2, "buy", pair, 10, pr, bufErrChan)

		for i := 0; i < cap(bufErrChan); i++ {
			if err := <-bufErrChan; err != nil {
				logging.Errorf("Error placing and filling: %s", err)
			}
			numRuns++
		}
	}
	return
}
