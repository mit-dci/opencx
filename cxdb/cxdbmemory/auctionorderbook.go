package cxdbmemory

import (
	"fmt"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// MemoryAuctionOrderbook is the representation of a auction orderbook for SQL
type MemoryAuctionOrderbook struct {
	// TODO: implement this

	// this pair
	pair *match.Pair
}

// CreateAuctionOrderbook creates a auction orderbook based on a pair
func CreateAuctionOrderbook(pair *match.Pair) (book match.AuctionOrderbook, err error) {
	// Set values for auction engine
	mo := &MemoryAuctionOrderbook{
		pair: pair,
	}
	// We can connect, now set return
	book = mo
	return
}

// UpdateBookExec takes in an order execution and updates the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookExec(exec *match.OrderExecution) (err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// UpdateBookCancel takes in an order cancellation and updates the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookCancel(cancel *match.CancelledOrder) (err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// UpdateBookPlace takes in an order, ID, auction ID, and adds the order to the orderbook.
func (mo *MemoryAuctionOrderbook) UpdateBookPlace(auctionIDPair *match.AuctionOrderIDPair) (err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetOrder gets an order from an OrderID
func (mo *MemoryAuctionOrderbook) GetOrder(orderID *match.OrderID) (aucOrder *match.AuctionOrderIDPair, err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// CalculatePrice takes in a pair and returns the calculated price based on the orderbook.
func (mo *MemoryAuctionOrderbook) CalculatePrice(auctionID *match.AuctionID) (price float64, err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// GetOrdersForPubkey gets orders for a specific pubkey.
func (mo *MemoryAuctionOrderbook) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.AuctionOrderIDPair, err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// ViewAuctionOrderbook takes in a trading pair and returns the orderbook as a map
func (mo *MemoryAuctionOrderbook) ViewAuctionOrderBook() (book map[float64][]*match.AuctionOrderIDPair, err error) {
	// TODO: Implement
	logging.Fatalf("UNIMPLEMENTED!")
	return
}

// CreateAuctionOrderbookMap creates a map of pair to auction engine, given a list of pairs.
func CreateAuctionOrderbookMap(pairList []*match.Pair) (aucMap map[match.Pair]match.AuctionOrderbook, err error) {

	aucMap = make(map[match.Pair]match.AuctionOrderbook)
	var curAucEng match.AuctionOrderbook
	for _, pair := range pairList {
		if curAucEng, err = CreateAuctionOrderbook(pair); err != nil {
			err = fmt.Errorf("Error creating single auction engine while creating auction engine map: %s", err)
			return
		}
		aucMap[*pair] = curAucEng
	}

	return
}
