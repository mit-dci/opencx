package main

import (
	"log"
	"os"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// Let these be turned into config things at some point
var (
	defaultServer = "hubris.media.mit.edu"
	defaultPort   = 12345
)

type openCxClient struct {
	RPCClient *cxrpc.OpencxRPCClient
}

// opencx-cli is the client, opencx is the server
func main() {
	var err error

	logging.SetLogLevel(2)

	commandArg := os.Args[1:]

	client := new(openCxClient)
	err = client.setupCxClient(defaultServer, defaultPort)

	if err != nil {
		log.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
	}

	// TODO just for now
	err = client.parseCommands(commandArg)
	if err != nil {
		log.Fatalf("Error parsing commands: \n%s", err)
	}
}

// NewOpenCxClient creates a new openCxClient for use as an RPC Client
func (cl *openCxClient) setupCxClient(server string, port int) error {
	var err error

	cl.RPCClient = new(cxrpc.OpencxRPCClient)

	err = cl.RPCClient.SetupConnection(server, port)
	if err != nil {
		return err
	}

	return nil
}

func (cl *openCxClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}
