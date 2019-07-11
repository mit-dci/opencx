package cxbenchmark

import (
	"fmt"
	"testing"

	"github.com/mit-dci/opencx/benchclient"
)

// BenchmarkPlaceOrders places orders on an unauthenticated server with full capabilities (besides auction)
func BenchmarkPlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start")

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var offChan chan bool
	if client1, client2, offChan, err = setupBenchmarkDualClient(false); err != nil {
		b.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	b.Logf("Test started client")
	// ugh when we run this benchmark and the server is noise then we basically crash the rpc server... need to figure out how to have that not happen, why is that fatal?
	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Logf("Running %s", placeFillTitle)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Logf("Running %s", placeBuySellTitle)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}

	b.Logf("stop benchmark rpc")
	offChan <- true

	return
}

// BenchmarkNoisePlaceOrders places orders on a NOISE-authenticated server with full capabilities (besides auction)
func BenchmarkNoisePlaceOrders(b *testing.B) {
	var err error

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var offChan chan bool
	if client1, client2, offChan, err = setupBenchmarkDualClient(true); err != nil {
		b.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Logf("Running %s", placeFillTitle)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Logf("Running %s", placeBuySellTitle)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}

	b.Logf("stop benchmark rpc")
	offChan <- true

	return
}

// BenchmarkEasyPlaceOrders places orders on an unauthenticated server with "pinky swear" settlement and full matching capabilites.
// This benchmark does standard, non auction matching.
func BenchmarkEasyPlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start")

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var offChan chan bool
	if client1, client2, offChan, err = setupBenchmarkDualClient(false); err != nil {
		b.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	b.Logf("Test started client")
	// ugh when we run this benchmark and the server is noise then we basically crash the rpc server... need to figure out how to have that not happen, why is that fatal?
	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Logf("Running %s", placeFillTitle)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Logf("Running %s", placeBuySellTitle)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}

	b.Logf("stop benchmark rpc")
	offChan <- true

	return
}

// BenchmarkEasyNoisePlaceOrders places orders on a NOISE-authenticated server with "pinky swear" settlement and full matching capabilites.
// This benchmark does standard, non auction matching.
func BenchmarkEasyNoisePlaceOrders(b *testing.B) {
	var err error

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var offChan chan bool
	if client1, client2, offChan, err = setupEasyBenchmarkDualClient(true); err != nil {
		b.Fatalf("Could not start dual client benchmark: \n%s", err)
	}

	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("PlaceAndFill%d", varRuns)
		b.Logf("Running %s", placeFillTitle)
		b.Run(placeFillTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceFillX(client1, client2, varRuns)
			}
		})
		placeBuySellTitle := fmt.Sprintf("PlaceBuyThenSell%d", varRuns)
		b.Logf("Running %s", placeBuySellTitle)
		b.Run(placeBuySellTitle, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PlaceBuySellX(client1, varRuns)
			}
		})
	}

	b.Logf("stop benchmark rpc")
	offChan <- true

	return
}
