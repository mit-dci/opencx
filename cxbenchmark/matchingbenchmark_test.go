package cxbenchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/cxrpc"
)

// func TestSetupDualClientNoAuth(t *testing.T) {
// 	var err error

// 	t.Logf("Start noauth server -- time: %s", time.Now())

// 	var offChan chan bool
// 	if _, _, offChan, err = setupBenchmarkDualClient(false); err != nil {
// 		offChan <- true
// 		t.Logf("Could not start noauth server: %s", err)
// 		return
// 	}

// 	t.Logf("Server noauth started -- time: %s", time.Now())
// 	close(offChan)
// 	return
// }

// func TestSetupDualClientAuth(t *testing.T) {
// 	var err error

// 	t.Logf("Start auth server -- time: %s", time.Now())

// 	var offChan chan bool
// 	if _, _, offChan, err = setupBenchmarkDualClient(true); err != nil {
// 		offChan <- true
// 		t.Logf("Could not start auth server: %s", err)
// 		return
// 	}

// 	t.Logf("Server auth started -- time: %s", time.Now())
// 	close(offChan)
// 	return
// }

// BenchmarkPlaceOrders places orders on an unauthenticated server with full capabilities (besides auction)
func BenchmarkPlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start -- time: %s", time.Now())

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var rpcListener *cxrpc.OpencxRPCCaller
	if client1, client2, rpcListener, err = setupBenchmarkDualClient(false); err != nil {
		b.Fatalf("Something is wrong with test, stopping benchmark")
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
	if err = rpcListener.Stop(); err != nil {
		err = fmt.Errorf("Error stopping rpc listener: %s", err)
		return
	}

	return
}

// BenchmarkNoisePlaceOrders places orders on a NOISE-authenticated server with full capabilities (besides auction)
func BenchmarkNoisePlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start -- time: %s", time.Now())

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var rpcListener *cxrpc.OpencxRPCCaller
	if client1, client2, rpcListener, err = setupBenchmarkDualClient(true); err != nil {
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
	if err = rpcListener.Stop(); err != nil {
		err = fmt.Errorf("Error stopping rpc listener: %s", err)
		return
	}

	return
}

// BenchmarkEasyPlaceOrders places orders on an unauthenticated server with "pinky swear" settlement and full matching capabilites.
// This benchmark does standard, non auction matching.
func BenchmarkEasyPlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start -- time: %s", time.Now())

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var rpcListener *cxrpc.OpencxRPCCaller
	if client1, client2, rpcListener, err = setupEasyBenchmarkDualClient(false); err != nil {
		b.Fatalf("Something is wrong with test, stopping benchmark: %s", err)
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
	if err = rpcListener.Stop(); err != nil {
		b.Errorf("Error stopping rpc listener for benchmark: %s", err)
	}

	return
}

// BenchmarkEasyNoisePlaceOrders places orders on a NOISE-authenticated server with "pinky swear" settlement and full matching capabilites.
// This benchmark does standard, non auction matching.
func BenchmarkEasyNoisePlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start -- time: %s", time.Now())

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var rpcListener *cxrpc.OpencxRPCCaller
	if client1, client2, rpcListener, err = setupEasyBenchmarkDualClient(true); err != nil {
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
	if err = rpcListener.Stop(); err != nil {
		err = fmt.Errorf("Error stopping rpc listener: %s", err)
		return
	}

	return
}
