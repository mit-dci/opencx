package cxserver

import "github.com/mit-dci/opencx/db/ocxsql"

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB   *ocxsql.DB
	OpencxRoot string
	OpencxPort int
	// TODO: Put TLS stuff here
}
