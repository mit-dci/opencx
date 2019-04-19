package cxbenchmark

import (
	"fmt"
	"testing"

	"github.com/mit-dci/opencx/cmd/benchclient"
	"github.com/mit-dci/opencx/logging"
)

// TODO: fix problems where you do something unexpected network-wise and the entire rpc server crashes!
func BenchmarkPlaceOrders(b *testing.B) {
	var err error

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	if client1, client2, err = setupBenchmarkDualClient(false); err != nil {
		logging.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	// ugh when we run this benchmark and the server is noise then we basically crash the rpc server... need to figure out how to have that not happen, why is that fatal?
	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}
}

func BenchmarkNoisePlaceOrders(b *testing.B) {
	var err error

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	if client1, client2, err = setupBenchmarkDualClient(true); err != nil {
		logging.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	if err = registerClient(client1); err != nil {
		logging.Fatalf("Could not register client2: \n%s", err)
	}

	if err = registerClient(client2); err != nil {
		logging.Fatalf("Could not register client2: \n%s", err)
	}

	runs := []int{}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}
}
