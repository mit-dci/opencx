package cxauctionserver

import (
	"runtime"
	"time"

	"github.com/mit-dci/opencx/logging"
)

// AuctionClock should be run in a goroutine and just commit to puzzles after some time
func (s *OpencxAuctionServer) AuctionClock() {
	logging.Infof("Starting Auction Clock!")

	// We make the variables here because we don't want to fill up our memory with stuff in the loop
	doneChan := make(chan time.Time, 1)
	var tickDone time.Time

	// afterTick is how we call the auction tick
	afterTick := func() {
		s.auctionTick(doneChan)
	}

	// FOR STATS / DEBUG
	var m *runtime.MemStats
	m = new(runtime.MemStats)

	for {
		// Read mem stats FOR STATISTICS
		runtime.ReadMemStats(m)

		// TODO: configurable time, work out schedule, base it on the AuctionTime option
		time.AfterFunc(time.Duration(s.t)*time.Microsecond, afterTick)

		// retrieve the tick from the channel
		tickDone = <-doneChan

		logging.Infof("Tick done at %s", tickDone.String())
	}
}

// auctionTick commits to orders and creates a new auction, while making sure to send a "done" time to a channel afterwards
func (s *OpencxAuctionServer) auctionTick(doneChan chan time.Time) {
	var err error

	// this basically makes sure we send something to doneChan
	// when we're done
	defer func() {
		doneChan <- time.Now()
	}()
	for pair, _ := range s.PuzzleEngines {
		if err = s.CommitOrdersNewAuction(&pair); err != nil {
			// TODO: What should happen in this case? How can we prevent this case?
			logging.Fatalf("Exchange commitment failed!!! Fatal error: %s", err)
		}
	}

	return
}
