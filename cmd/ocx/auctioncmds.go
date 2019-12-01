package main

import (
	"fmt"
	"strconv"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/Rjected/lit/lnutil"
	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

var placeAuctionOrderCommand = &Command{
	Format: fmt.Sprintf("%s%s%s%s%s\n", lnutil.Red("placeauctionorder"), lnutil.ReqColor("side"), lnutil.ReqColor("pair"), lnutil.ReqColor("amounthave"), lnutil.ReqColor("price")),
	Description: fmt.Sprintf("%s\n%s\n",
		"Submit a front-running resistant auction order with side \"buy\" or side \"sell\", for pair \"asset1\"/\"asset2\", where you give up amounthave of \"asset1\" (if on buy side) or \"asset2\" if on sell side, for the other token at a specific price.",
		"This will return an order ID which can be used as input to cancelorder, or getorder.",
	),
	ShortDescription: fmt.Sprintf("%s\n", "Place a front-running resistant order on the exchange."),
}

// OrderCommand submits an order (for now)
func (cl *ocxClient) AuctionOrderCommand(args []string) (err error) {
	if err = cl.UnlockKey(); err != nil {
		logging.Fatalf("Could not unlock key! Fatal!")
	}

	side := args[0]
	pair := args[1]

	amountHave, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		err = fmt.Errorf("Error parsing amountHave, please enter something valid:\n%s", err)
		return
	}

	price, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		err = fmt.Errorf("Error parsing price: \n%s", err)
		return
	}

	var pubkey *koblitz.PublicKey
	if pubkey, err = cl.RetrievePublicKey(); err != nil {
		return
	}

	pairParam := new(match.Pair)
	if err = pairParam.FromString(pair); err != nil {
		err = fmt.Errorf("Error parsing pair, please enter something valid: %s", err)
		return
	}

	var paramreply *cxauctionrpc.GetPublicParametersReply
	if paramreply, err = cl.RPCClient.GetPublicParameters(pairParam); err != nil {
		err = fmt.Errorf("Error getting public parameters before placing auction order: %s", err)
		return
	}

	// we ignore reply because there's nothing in it and we don't use it
	// var reply *cxauctionrpc.SubmitPuzzledOrderReply
	if _, err = cl.RPCClient.AuctionOrderCommand(pubkey, side, pair, amountHave, price, paramreply.AuctionTime, paramreply.AuctionID); err != nil {
		return
	}

	logging.Infof("Successfully placed auction order")

	return
}
