package cxserver

import (
	"fmt"

	"github.com/mit-dci/opencx/match"
)

func (server *OpencxServer) GetPrice(pair *match.Pair) (price float64, err error) {
	server.dbLock.Lock()
	var currOrderbook match.LimitOrderbook
	var ok bool
	if currOrderbook, ok = server.Orderbooks[*pair]; !ok {
		err = fmt.Errorf("Could not find orderbooks for trading pair for GetPrice")
		server.dbLock.Unlock()
		return
	}

	if price, err = currOrderbook.CalculatePrice(); err != nil {
		err = fmt.Errorf("Error calculating price for server GetPrice: %s", err)
		server.dbLock.Unlock()
		return
	}
	server.dbLock.Unlock()
	return
}
