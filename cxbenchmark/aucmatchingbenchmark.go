package cxbenchmark

import (
	"github.com/mit-dci/opencx/benchclient"
	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// AuctionPlaceAndFill places a whole bunch of orders (in goroutines of course) and fills a whole bunch of orders.
func AuctionPlaceAndFill(client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, pair string, howMany int) {

	// get pair from the string
	pairParam := new(match.Pair)
	if err := pairParam.FromString(pair); err != nil {
		logging.Errorf("Error getting pair from string for AuctionPlaceAndFill: %s", err)
		return
	}

	for i := 0; i < howMany; i++ {
		// This shouldnt make any change in balance but each account should have at least 2000 satoshis (or the smallest unit in whatever chain)
		bufErrChan := make(chan error, 4)
		orderChan := make(chan *cxauctionrpc.SubmitPuzzledOrderReply)

		publicParams, err := client1.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing and filling auction: %s", err)
		}
		go client1.AuctionOrderAsync(client1.PrivKey.PubKey(), "buy", pair, 1000, 1.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)
		publicParams, err = client2.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing and filling auction: %s", err)
		}
		go client2.AuctionOrderAsync(client2.PrivKey.PubKey(), "sell", pair, 1000, 1.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)
		publicParams, err = client1.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing and filling auction: %s", err)
		}
		go client1.AuctionOrderAsync(client1.PrivKey.PubKey(), "sell", pair, 2000, 2.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)
		publicParams, err = client2.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing and filling auction: %s", err)
		}
		go client2.AuctionOrderAsync(client2.PrivKey.PubKey(), "buy", pair, 1000, 2.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)

		for i := 0; i < cap(bufErrChan); i++ {
			select {
			case err := <-bufErrChan:
				if err != nil {
					logging.Errorf("Error placing and filling auction: %s", err)
				}
			case order := <-orderChan:
				if order != nil {
					logging.Infof("Placed puzzled order - place and fill")
				}
			}
			if err := <-bufErrChan; err != nil {
				logging.Errorf("Error placing and filling auction: %s", err)
			}
		}
	}
	return
}

// AuctionPlaceManyBuy places many orders at once
func AuctionPlaceManyBuy(client *benchclient.BenchClient, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxauctionrpc.SubmitPuzzledOrderReply)

	// get pair from the string
	pairParam := new(match.Pair)
	if err := pairParam.FromString(pair); err != nil {
		logging.Errorf("Error getting pair from string for AuctionPlaceManyBuy: %s", err)
		return
	}

	for i := 0; i < howMany; i++ {
		publicParams, err := client.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing many buy auction: %s", err)
		}
		go client.AuctionOrderAsync(client.PrivKey.PubKey(), "buy", pair, 1000, 1.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)
	}

	for i := 0; i < cap(bufErrChan); i++ {
		select {
		case err := <-bufErrChan:
			if err != nil {
				logging.Errorf("Error placing many: %s", err)
			}
		case order := <-orderChan:
			if order != nil {
				logging.Infof("Placed puzzled order - buy")
			}
		}
	}

	return
}

// AuctionPlaceManySell places many orders at once
func AuctionPlaceManySell(client *benchclient.BenchClient, pair string, howMany int) {
	bufErrChan := make(chan error, howMany)
	orderChan := make(chan *cxauctionrpc.SubmitPuzzledOrderReply)

	// get pair from the string
	pairParam := new(match.Pair)
	if err := pairParam.FromString(pair); err != nil {
		logging.Errorf("Error getting pair from string for AuctionPlaceManySell: %s", err)
		return
	}

	for i := 0; i < howMany; i++ {
		publicParams, err := client.GetPublicParameters(pairParam)
		if err != nil {
			logging.Errorf("Error placing many sell auction: %s", err)
		}
		go client.AuctionOrderAsync(client.PrivKey.PubKey(), "sell", pair, 1000, 1.0, publicParams.AuctionTime, publicParams.AuctionID, orderChan, bufErrChan)
	}

	for i := 0; i < cap(bufErrChan); i++ {
		select {
		case err := <-bufErrChan:
			if err != nil {
				logging.Errorf("Error placing many: %s", err)
			}
		case order := <-orderChan:
			if order != nil {
				logging.Infof("Placed puzzled order - sell")
			}
		}
	}

	return
}

// AuctionPlaceBuySellX runs functions with predone parameters, you only tell it how many times it should run and what client to use
func AuctionPlaceBuySellX(client *benchclient.BenchClient, varRuns int) {
	AuctionPlaceManyBuy(client, "regtest/litereg", varRuns)
	// AuctionPlaceManyBuy(client, "regtest/vtcreg", varRuns)
	// AuctionPlaceManyBuy(client, "vtcreg/litereg", varRuns)
	AuctionPlaceManySell(client, "regtest/litereg", varRuns)
	// AuctionPlaceManySell(client, "regtest/vtcreg", varRuns)
	// AuctionPlaceManySell(client, "vtcreg/litereg", varRuns)
	return
}

// AuctionPlaceFillX runs functions with predone parameters, you only tell it how many times it should run and what client to use
func AuctionPlaceFillX(client1 *benchclient.BenchClient, client2 *benchclient.BenchClient, varRuns int) {
	AuctionPlaceAndFill(client1, client2, "regtest/litereg", varRuns)
	// AuctionPlaceAndFill(client1, client2, "regtest/vtcreg", varRuns)
	// AuctionPlaceAndFill(client1, client2, "vtcreg/litereg", varRuns)
}
