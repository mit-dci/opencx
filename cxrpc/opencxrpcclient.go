package cxrpc

import (
	"fmt"
	"io"
	"log"
	"net/rpc"
)

// OpencxRPCClient is a RPC client for the opencx server
type OpencxRPCClient struct {
	Conn   *rpc.Client
	token  []byte
	logger *log.Logger
}

// Call calls the servicemethod with name stirng, args args, and reply reply
func(cl *OpencxRPCClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.Conn.Call(serviceMethod, args, reply)
}

// SetupConnection creates a new RPC client
func(cl *OpencxRPCClient) SetupConnection(server string, port int) error {
	var err error

	cl.Conn, err = rpc.Dial("tcp", server + ":" + fmt.Sprintf("%d",port))
	if err != nil {
		return err
	}

	cl.Printf("Connected to exchange at %s\n", server + ":" + fmt.Sprintf("%d", port))
	return nil
}

// SetupLogger sets up the client side logging
func(cl *OpencxRPCClient) SetupLogger(w io.Writer) {
	cl.logger = log.New(w, "OCX Client: ", log.LstdFlags)
	cl.logger.Println("Set up logger")
}

// Printf is actually needed so we can print without worrying whether or not the logger is set, but print to the logger when we have one
func(cl *OpencxRPCClient) Printf(format string, v ...interface{}) {
	if cl.logger != nil {
		cl.logger.Printf(format, v...)
	}
}
