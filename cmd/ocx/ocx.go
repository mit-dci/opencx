package main

import (
	"log"
	"os"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/logging"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = 12345
)

// TODO figure out this, call in functions specific to method
type openCxClient struct {
	Username  string
	Token     []byte
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

	// TODO: change to file logging when it's needed, not now
	cl.RPCClient.SetupLogger(os.Stdout)
	if err != nil {
		return err
	}

	err = cl.RPCClient.SetupConnection(server, port)
	if err != nil {
		return err
	}

	return nil
}

func (cl *openCxClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}

func (cl *openCxClient) Printf(format string, v ...interface{}) {
	cl.RPCClient.Printf(format, v...)
}
