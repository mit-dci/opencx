package main

import (
	"fmt"
)

func(cl *openCxClient) parseCommands(commands []string) error {
	var args []string

	if len(commands) == 0 {
		return fmt.Errorf("Please specify arguments for exchange CLI")
	}
	cmd := commands[0]

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
		err := cl.Register(args)
		if err != nil {
			return err
		}
		// method that uses rpc should also set token to instance of client if returned
	}
	if cmd == "login" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify two arguments: username and password. Instead, %d arguments were specified", len(args))
		}

		err := cl.Login(args)
		if err != nil {
			return fmt.Errorf("Error calling login when parsing:\n%s", err)
		}
	}
	if cmd == "vieworderbook" {
		// run method that returns orders in json
	}
	return nil
}
