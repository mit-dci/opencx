package main

import (
	"log"
	"net"
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

	defer listener.Close()
	rpc.Accept(listener)

}
