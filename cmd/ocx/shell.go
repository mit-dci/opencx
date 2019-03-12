package main

import (
	"github.com/fatih/color"
	"github.com/mit-dci/lit/lnutil"

	"fmt"
)

// Command holds information about commands so we can show all the stuff you can do
type Command struct {
	Format           string
	Description      string
	ShortDescription string
}

var helpCommand = &Command{
	Format:           fmt.Sprintf("%s%s\n", lnutil.White("help"), lnutil.OptColor("command")),
	Description:      "Show information about a given command\n",
	ShortDescription: "Show information about a given command\n",
}

func (cl *openCxClient) parseCommands(commands []string) error {
	var args []string

	if len(commands) == 0 {
		return fmt.Errorf("Please specify arguments for exchange CLI")
	}
	cmd := commands[0]

	if len(commands) > 1 {
		args = commands[1:]
	}
	// help gives you really terse help.  Just a list of commands.
	if cmd == "help" {
		err := cl.Help(args)
		return err
	}
	if cmd == "register" {
		if len(args) != 0 {
			return fmt.Errorf("Please do not specify any arguments. You do not need a username, you will be registered by public key")
		}

		if err := cl.Register(args); err != nil {
			return err
		}
	}
	if cmd == "getbalance" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify token to get balance for token")
		}

		if err := cl.GetBalance(args); err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getallbalances" {
		if len(args) != 0 {
			return fmt.Errorf("Please do not specify any arguments")
		}

		if err := cl.GetAllBalances(args); err != nil {
			return fmt.Errorf("Error getting balance: \n%s", err)
		}
	}
	if cmd == "getdepositaddress" {
		if len(args) != 1 {
			return fmt.Errorf("Must specify asset to get deposit address for asset")
		}

		if err := cl.GetDepositAddress(args); err != nil {
			return fmt.Errorf("Error getting deposit address: \n%s", err)
		}
	}
	if cmd == "placeorder" {
		if len(args) != 4 {
			return fmt.Errorf("Must specify 4 arguments: side, pair, amountHave, and price")
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
		if len(args) != 3 {
			return fmt.Errorf("Must specify 3 arguments: amount coin address")
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
		if len(args) != 0 {
			return fmt.Errorf("Don't specify arguments please")
		}

		if err := cl.GetLitConnection(args); err != nil {
			return fmt.Errorf("Error getting lit connection: \n%s", err)
		}
	}
	return nil
}

func printHelp(commands []*Command) {
	for _, command := range commands {
		fmt.Fprintf(color.Output, "%s\t%s", command.Format, command.ShortDescription)
	}
}

func (cl *openCxClient) Help(textArgs []string) error {
	if len(textArgs) == 0 {

		fmt.Fprintf(color.Output, lnutil.Header("Commands:\n"))
		listofCommands := []*Command{}
		printHelp(listofCommands)
		return nil
	}

	if textArgs[0] == "help" || textArgs[0] == "-h" {
		fmt.Fprintf(color.Output, helpCommand.Format)
		fmt.Fprintf(color.Output, helpCommand.Description)
		return nil
	}
	res := make([]string, 0)
	res = append(res, textArgs[0])
	res = append(res, "-h")
	return cl.parseCommands(res)
}
