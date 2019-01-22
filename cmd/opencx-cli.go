package cmd

import (
	"fmt"
	"os"
)

// opencx-cli is the client, opencx is the server
func main() {
	commandArg := os.Args[1:]

	// TODO just for now
	println(commandArg)
}

func parseCommands(commands []string) error {
	var args []string
	cmd := args[0]
	if len(commands) > 1 {
		args = commands[1:]
	}
	if cmd == "register" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify two arguments: username and password. Instead, %d arguments were specified", len(args))
		}
		username := args[0]
		password := args[1]
		// construct JSON and send through rpc
		err := OpencxRPCClient.Register(username, password)
		// method that uses rpc should also set token to instance of client if returned
	}
	if cmd == "login" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify two arguments: username and password. Instead, %d arguments were specified", len(args))
		}
	}
	if cmd == "vieworderbook" {
		// run method that
	}
	return nil
}
