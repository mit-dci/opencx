package cxserver

import (
	"fmt"

	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/wallit"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB     *ocxsql.DB
	OpencxRoot   string
	OpencxPort   int
	OpencxPubkey interface{}
	// TODO: Put TLS stuff here
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}

// NewChildAddress creates a child address, to be assigned to an account, from the current pubkey
func (server *OpencxServer) NewChildAddress() error {
	rootKey := new(hdkeychain.ExtendedKey)
	params := new(coinparam.Params)

	birthHeight := int32(0)
	resync := true
	spvhost := ""
	path := ""
	proxyURL := ""

	wallit, _, err := wallit.NewWallit(rootKey, birthHeight, resync, spvhost, path, proxyURL, params)
	if err != nil {
		return fmt.Errorf("Error when creating wallit: \n%s", err)
	}

	fmt.Printf("Ay address: %s\n", wallit.GetWalletPrivkey(uint32(132)))

	return nil
}
