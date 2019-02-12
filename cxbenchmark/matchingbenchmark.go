package cxbenchmark

import (
	"github.com/mit-dci/opencx/cmd/benchclient"
)

// PlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func PlaceAndFill(client *benchclient.BenchClient, user1 string, user2 string, pair string, howMany int) {
	for i := 0; i < howMany; i++ {
		for pr := 0.0; pr < 5; pr = pr + 0.1 {
			// This shouldnt make any change in balance but each account should have at least 500 satoshis (of the smallest unit in whatever chain)
			go client.OrderCommand(user1, "buy", pair, 10, pr)
			go client.OrderCommand(user2, "sell", pair, 10, pr)
			go client.OrderCommand(user1, "sell", pair, 10, pr)
			go client.OrderCommand(user2, "buy", pair, 10, pr)
		}
	}
}
