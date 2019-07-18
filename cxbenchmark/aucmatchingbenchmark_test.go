package cxbenchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/cxauctionrpc"
)

func TestSetupDualClientNoAuth(t *testing.T) {
	var err error

	t.Logf("Start noauth server -- time: %s", time.Now())

	var rpcListener *cxauctionrpc.AuctionRPCCaller
	if _, _, rpcListener, err = setupEasyAuctionBenchmarkDualClient(false); err != nil {
		t.Errorf("Could not start noauth server: %s", err)
		return
	}

	t.Logf("Server noauth started -- time: %s", time.Now())
	if err = rpcListener.KillServerNoWait(); err != nil {
		t.Errorf("Error killing server for TestSetupDualClientNoAuth: %s", err)
		return
	}
	return
}

// BenchmarkEasyAuctionPlaceOrders places orders on an unauthenticated auction server with "pinky swear" settlement and full matching capabilites.
func BenchmarkEasyAuctionPlaceOrders(b *testing.B) {
	var err error

	b.Logf("Test start -- time: %s", time.Now())

	var client1 *benchclient.BenchClient
	var client2 *benchclient.BenchClient
	var rpcListener *cxauctionrpc.AuctionRPCCaller
	if client1, client2, rpcListener, err = setupEasyAuctionBenchmarkDualClient(false); err != nil {
		b.Errorf("Could not start dual client benchmark: \n%s", err)
		return
	}

	b.Logf("Test started client - %s", time.Now())
	// ugh when we run this benchmark and the server is noise then we basically crash the rpc server... need to figure out how to have that not happen, why is that fatal?
	runs := []int{1}
	for _, varRuns := range runs {
		placeFillTitle := fmt.Sprintf("AuctionPlaceAndFill%d", varRuns)
		b.Logf("Running %s", placeFillTitle)
		AuctionPlaceFillX(client1, client2, varRuns)
		placeBuySellTitle := fmt.Sprintf("AuctionPlaceBuyThenSell%d", varRuns)
		b.Logf("Running %s", placeBuySellTitle)
		AuctionPlaceBuySellX(client1, varRuns)
	}
	b.Logf("waiting for results - %s", time.Now())
	if err = rpcListener.KillServerWait(); err != nil {
		b.Errorf("Error killing server for BenchmarkEasyAuctionPlaceOrders: %s", err)
		return
	}

	return
}
