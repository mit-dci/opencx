package cxdbmemory

import (
	"fmt"
	"sync"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

type MemoryAuctionEngine struct {
	orders     map[match.AuctionID]map[float64][]*match.AuctionOrderIDPair
	auctionMtx *sync.Mutex
	pair       *match.Pair
}

// PlaceAuctionOrder should place an order for a specific auction ID, and produce a response output.
// This response output should be used in case the matching engine dies, and this can be replayed to build the state.
// This method assumes that the auction order is valid, and has the same pair as all of the other orders that have been placed for this matching engine.
func (me *MemoryAuctionEngine) PlaceAuctionOrder(order *match.AuctionOrder, auctionID *match.AuctionID) (idRes *match.AuctionOrderIDPair, err error) {
	me.auctionMtx.Lock()

	idCopy := *auctionID

	// First get the price of the order, if this errors then that's really bad
	var pr float64
	if pr, err = order.Price(); err != nil {
		err = fmt.Errorf("Critical error when placing order for matching engine: %s", err)
		me.auctionMtx.Unlock()
		return
	}

	// Now create an ID. If this errors then that's really bad
	var id [32]byte
	hasher := sha3.New256()
	hasher.Write(order.SerializeSignable())
	copy(id[:], hasher.Sum(nil))

	idRes = &match.AuctionOrderIDPair{
		OrderID: id,
		Order:   order,
	}

	// We assume that the order has been properly validated when it goes in to the auction orderbook
	var ok bool
	// If the map for the auction isn't there, create it
	if _, ok = me.orders[idCopy]; !ok {

		// Since we assume the order is valid, place it in the auction
		me.orders[idCopy] = map[float64][]*match.AuctionOrderIDPair{
			pr: []*match.AuctionOrderIDPair{
				idRes,
			},
		}
		return
	}

	// if the map for the auction is there but the price index isn't, create it
	if _, ok = me.orders[idCopy][pr]; !ok {
		me.orders[idCopy][pr] = []*match.AuctionOrderIDPair{
			idRes,
		}
		return
	}

	// if both are fine then awesome
	me.orders[idCopy][pr] = append(me.orders[idCopy][pr], idRes)

	me.auctionMtx.Unlock()
	// TODO
	return
}

// CancelAuctionOrder should cancel an order for a specific order ID, and produce a response output.
// This response output should be used in case the matching engine dies, and this can be replayed to build the state.
func (me *MemoryAuctionEngine) CancelAuctionOrder(id *match.OrderID) (cancelled *match.CancelledOrder, cancelSettlement *match.SettlementExecution, err error) {
	me.auctionMtx.Lock()
	// Go through the maps and, since we don't have an order ID => order index just delete em all
	var deletedOrder *match.AuctionOrderIDPair
	for _, orderMap := range me.orders {
		for pr, orderIDPairList := range orderMap {
			var deletedIdx int
			var deleted bool
			for idx, orderIDPair := range orderIDPairList {
				if orderIDPair.OrderID == *id {
					deletedOrder = orderIDPair
					deleted = true
					deletedIdx = idx
				}
			}

			// If deleted remove from thing
			if deleted {
				oidLen := len(orderIDPairList)
				orderIDPairList[oidLen-1], orderIDPairList[deletedIdx] = orderIDPairList[deletedIdx], orderIDPairList[oidLen-1]
				orderMap[pr] = orderIDPairList[:oidLen-1]
			}
		}
	}
	// Get side from string rip
	var orderSide *match.Side
	orderSide = new(match.Side)
	if err = orderSide.FromString(deletedOrder.Order.Side); err != nil {
		err = fmt.Errorf("Error getting order side from string for CancelOrder: %s", err)
		me.auctionMtx.Unlock()
		return
	}

	var debitAsset match.Asset
	if *orderSide == match.Buy {
		debitAsset = me.pair.AssetHave
	} else {
		debitAsset = me.pair.AssetWant
	}
	cancelled = &match.CancelledOrder{
		OrderID: id,
	}

	cancelSettlement = &match.SettlementExecution{
		Pubkey: deletedOrder.Order.Pubkey,
		Amount: deletedOrder.Order.AmountHave,
		Asset:  debitAsset,
		Type:   match.Debit,
	}
	me.auctionMtx.Unlock()
	return
}

// MatchAuctionOrders matches the auction orders for a specific auction ID
func (me *MemoryAuctionEngine) MatchAuctionOrders(auctionID *match.AuctionID) (orderExecs []*match.OrderExecution, settlementExecs []*match.SettlementExecution, err error) {
	// TODO
	return
}

func CreateAuctionEngineMap(coinList []*coinparam.Params) (mengines map[match.Pair]match.AuctionEngine, err error) {
	mengines = make(map[match.Pair]match.AuctionEngine)

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		err = fmt.Errorf("Error generating asset pairs while creating auction engine map: %s", err)
		return
	}

	// We just create a new struct because that's all we really need, we satisfy the interface
	for _, pair := range pairList {
		mengines[*pair] = new(MemoryAuctionEngine)
	}

	return
}
