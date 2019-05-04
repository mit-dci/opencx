package cxauctionserver

import (
	"time"

	"github.com/mit-dci/opencx/logging"
)

// AuctionClock should be run in a goroutine and just commit to puzzles after some time
func (s *OpencxAuctionServer) AuctionClock() {
	logging.Infof("Starting Auction Clock!")
	for {
		logging.Infof("Auction clock tick!")

		// TODO: configurable time, work out schedule, base it on the AuctionTime option
		doneChan := make(chan time.Time, 1)
		time.AfterFunc(time.Second, func() {
			var err error

			// this basically makes sure we send something to doneChan
			// when we're done
			defer func() {
				doneChan <- time.Now()
			}()
			if err = s.CommitOrdersNewAuction(); err != nil {
				// TODO: What should happen in this case? How can we prevent this case?
				logging.Fatalf("Exchange commitment failed!!! Fatal error: %s", err)
			}
			return
		})

		logging.Infof("Waiting for tick")
		tickDone := <-doneChan

		logging.Infof("Tick done at %s", tickDone.String())
		close(doneChan)
	}
}
