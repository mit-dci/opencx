package cxserver

import (

	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB      *ocxsql.DB
	OpencxRoot    string
	OpencxPort    int
	OpencxPrivkey *hdkeychain.ExtendedKey
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}
