package main

import (
	"log"
	"fmt"
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

	listener, err := net.Listen("tcp", ":12345")
	fmt.Printf("Running server on %s\n", listener.Addr().String())
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer listener.Close()
	rpc.Accept(listener)

}
