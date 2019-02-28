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
		if len(args) != 1 {
			return fmt.Errorf("Must specify one argument: username. Instead, %d arguments were specified", len(args))
		}

		if err := cl.Register(args); err != nil {
			return err
		}
	}
	if cmd == "getbalance" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify username and token to get balance for token")
		}

		if err := cl.GetBalance(args); err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getallbalances" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify username to get balances for user")
		}

		if err := cl.GetAllBalances(args); err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getdepositaddress" {
		if len(args) != 2 {
			return fmt.Errorf("Must specify username and asset to get deposit address for asset")
		}

		if err := cl.GetDepositAddress(args); err != nil {
			return fmt.Errorf("Error getting deposit address: \n%s", err)
		}
	}
	if cmd == "placeorder" {
		if len(args) != 5 {
			return fmt.Errorf("Must specify 5 arguments: name, side, pair, amountHave, and Price")
		}

		if err := cl.OrderCommand(args); err != nil {
			return fmt.Errorf("Error calling order command: \n%s", err)
		}
	}
	if cmd == "vieworderbook" {
		if len(args) != 1 || len(args) != 2 {
			return fmt.Errorf("Must specify from 1 to 2 arguments: pair [buy|sell]")
		}

		if err := cl.ViewOrderbook(args); err != nil {
			return fmt.Errorf("Error calling vieworderbook command: \n%s", err)
		}
	}
	if cmd == "getprice" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify 1 argument: pair")
		}

		if err := cl.GetPrice(args); err != nil {
			return fmt.Errorf("Error calling getprice command: \n%s", err)
		}
	}
	if cmd == "withdraw" {
		if len(args) != 4 {
			return fmt.Errorf("Must specify 4 arguments: name amount coin address")
		}

		if err := cl.Withdraw(args); err != nil {
			return fmt.Errorf("Error calling withdraw command: \n%s", err)
		}
	}
	if cmd == "cancelorder" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify 1 argument: orderID")
		}

		if err := cl.CancelOrder(args); err != nil {
			return fmt.Errorf("Error calling cancel command: \n%s", err)
		}
	}
	if cmd == "getpairs" {
		if len(args) != 0 {
			return fmt.Errorf("Don't specify arguments please")
		}

		if err := cl.GetPairs(); err != nil {
			return fmt.Errorf("Error getting pairs: \n%s", err)
		}
	}
	if cmd == "getlitconnection" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify one argument: asset")
		}

		if err := cl.GetLitConnection(args); err != nil {
			return fmt.Errorf("Error getting lit connection: \n%s", err)
		}
	}
	return nil
}
