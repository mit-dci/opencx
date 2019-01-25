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
		if len(args) != 2 {
			return fmt.Errorf("Must specify 2 currencies to view the pair's orderbook")
		}

		err := cl.ViewOrderbook(args[0],args[1])
		if err != nil {
			return fmt.Errorf("Error viewing orderbook: \n%s", err)
		}
	}
	if cmd == "nologinuseless" {
		if len(args) > 0 {
			cl.Username = args[0]
		} else {
			cl.Username = "fakeusername"
		}
		err := cl.AuthCommand(args, cl.uselessFunction)
		if err != nil {
			return fmt.Errorf("Error when calling useless function when not logged in:\n%s", err)
		}
	}
	if cmd == "loginuseless" {
		if len(args) < 2 {
			return fmt.Errorf("Must specify at least two arguments: username and password. Instead, %d arguments were specified", len(args))
		}

		err := cl.Login(args[0:2])
		if err != nil {
			return fmt.Errorf("Error calling login when parsing:\n%s", err)
		}

		err = cl.AuthCommand(args[2:], cl.uselessFunction)
		if err != nil {
			return fmt.Errorf("Error when calling useless function when logged in:\n%s", err)
		}
	}
	return nil
}
