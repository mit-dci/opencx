package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mit-dci/opencx/cxrpc"
)

var (
	defaultServer = "localhost"
	defaultPort   = 12345
)

// TODO figure out this, call in functions specific to method
type openCxClient struct {
	RPCClient *cxrpc.OpencxRPCClient
}

// opencx-cli is the client, opencx is the server
func main() {
	commandArg := os.Args[1:]

	// TODO just for now
	err := parseCommands(commandArg)
	if err != nil {
		log.Fatal(err)
	}
}

func parseCommands(commands []string) error {
	var args []string

	if len(commands) == 0 {
		return fmt.Errorf("Please specify arguments for exchange CLI")
	}
	cmd := commands[0]

	// TODO figure out if this is right
	client, err := cxrpc.NewOpencxRPCClient(defaultServer, defaultPort)
	if err != nil {
		return err
	}

	if len(commands) > 1 {
		args = commands[1:]
	}
	if cmd == "register" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify two arguments: username and password. Instead, %d arguments were specified", len(args))
		}

		// TODO call register method here, that method does rpc.Call with the appropriate stuff

		// construct JSON and send through rpc
		// call client register function with args
		err := client.Register(args)
		if err != nil {
			return err
		}
		// method that uses rpc should also set token to instance of client if returned
	}
	if cmd == "login" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify two arguments: username and password. Instead, %d arguments were specified", len(args))
		}
	}
	if cmd == "vieworderbook" {
		// run method that returns orders in json
	}
	return nil
}
