package main

import (
	"fmt"
)

func (cl *openCxClient) parseCommands(commands []string) error {
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
	if cmd == "getbalance" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify username and token to get balance for token")
		}

		err := cl.GetBalance(args)
		if err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getallbalances" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify username to get balances for user")
		}

		err := cl.GetAllBalances(args)
		if err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getdepositaddress" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify username and token to get deposit address for toke")
		}

		err := cl.GetDepositAddress(args)
		if err != nil {
			return fmt.Errorf("Error getting deposit address: \n%s", err)
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
	if cmd == "placeorder" {
		if len(args) != 5 {
			return fmt.Errorf("Must specify 5 arguments: name, side, pair, amountHave, and Price")
		}

		err := cl.OrderCommand(args)
		if err != nil {
			return fmt.Errorf("Error calling order command: \n%s", err)
		}
	}
	if cmd == "vieworderbook" {
		if len(args) > 2 {
			return fmt.Errorf("Must specify 2 or 1 arguments: pair [buy|sell]")
		}

		err := cl.ViewOrderbook(args)
		if err != nil {
			return fmt.Errorf("Error calling vieworderbook command: \n%s", err)
		}
	}
	if cmd == "getprice" {
		if len(args) > 1 {
			return fmt.Errorf("Must specify 1 argument: pair")
		}

		err := cl.GetPrice(args)
		if err != nil {
			return fmt.Errorf("Error calling getprice command: \n%s", err)
		}
	}
	return nil
}
