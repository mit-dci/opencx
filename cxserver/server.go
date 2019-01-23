package cxserver

import "github.com/mit-dci/opencx/db/ocxredis"

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB   *ocxredis.DB
	OpencxRoot string
	OpencxPort int
}
