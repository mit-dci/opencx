package cxauctionserver

import (
	"runtime"
	"time"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

type timeID struct {
	time time.Time
	id   [32]byte
}

// AuctionClock should be run in a goroutine and just commit to puzzles after some time
// What this does is first makes a channel called doneChan.
// This channel keeps track of the time that each auction is committed to.
func (s *OpencxAuctionServer) AuctionClock(pair match.Pair, startID [32]byte) {
	logging.Infof("Starting Auction Clock!")

	// We make the variables here because we don't want to fill up our memory with stuff in the loop
	doneChan := make(chan timeID, 1)
	var tickResult timeID

	// FOR STATS / DEBUG
	var m *runtime.MemStats
	m = new(runtime.MemStats)

	// This takes currAuctionID, plugs it in to auctionTick, gets back a new auction id from the channel so it
	// can continue the loop
	var currAuctionID [32]byte = startID
	for {
		// Read mem stats FOR STATISTICS
		runtime.ReadMemStats(m)

		// TODO: configurable time, work out schedule, base it on the AuctionTime option
		time.AfterFunc(time.Duration(s.t)*time.Microsecond, func() {
			s.auctionTick(pair, currAuctionID, doneChan)
		})

		// retrieve the tick from the channel
		tickResult = <-doneChan

		// Now set the new id to the result
		currAuctionID = tickResult.id

		logging.Infof("Tick done at %s", tickResult.time.String())
	}
}

// auctionTick commits to orders and creates a new auction, while making sure to send a "done" time to a channel afterwards
func (s *OpencxAuctionServer) auctionTick(pair match.Pair, oldID [32]byte, doneChan chan timeID) {
	var err error

	var tickRes timeID

	// this basically makes sure we send something to doneChan
	// when we're done
	defer func() {
		doneChan <- tickRes
	}()
	// batcher solves puzzles, puzzle engine stores puzzles.
	if tickRes.id, err = s.CommitOrdersNewAuction(&pair, oldID); err != nil {
		// TODO: What should happen in this case? How can we prevent this case?
		logging.Fatalf("Exchange commitment for %x failed!!! Fatal error: %s", oldID, err)
	}

	// Now set the time
	tickRes.time = time.Now()

	return
}
