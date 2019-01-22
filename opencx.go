package main

import (
	"net/rpc"

	"github.com/mit-dci/opencx/cxrpc"
)

func main() {

	rpc1 := new(cxrpc.OpencxRPC)
	err := rpc.Register(rpc1)
	if err != nil {
		println(err.Error())
	}
}
