package match

import (
	"time"

	"github.com/mit-dci/opencx/logging"
)

// LimitOrderIDPair is order ID, order, and time, used for generating executions in limit order matching algorithms
type LimitOrderIDPair struct {
	Timestamp time.Time
	OrderID   [32]byte
	Order     *LimitOrder
}

// GeneratePTPExecs goes through an orderbook and generates executions based on the Price/Time Priority
// matching algorithm (PTP)
func GeneratePTPExecs(book map[float64][]*LimitOrderIDPair) (executions []*OrderExecution, err error) {

	// Input: map from price to orders, you don't know which are buy or sell or anything
	var resExec *OrderExecution
	// Go through prices and figure out if there are any to match
	for price, _ := range book {
		logging.Infof("unimplemented!! %f %s", price, resExec.String())
	}

	return
}
