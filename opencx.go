package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/mit-dci/opencx/cxrpc"
)

func main() {

	rpc1 := new(cxrpc.OpencxRPC)
	err := rpc.Register(rpc1)
	if err != nil {
		log.Fatal("register error:", err)
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	go http.Serve(listener, nil)

}
