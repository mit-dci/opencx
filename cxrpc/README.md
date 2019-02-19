# cxrpc

This package handles RPC requests coming in to the exchange. Here are all the commands supported so far:
RPC is just a starting point for being able to accept network I/O

## register
Register registers an account if that username does not exist already

`ocx register name`

Arguments:
 - Name (string)

Outputs:
- A message that says you successfully registered or an error

## vieworderbook
Vieworderbook shows you the current orderbook

`ocx vieworderbook pair [buy/sell]`

Arguments:
 - Asset pair (string)
 - buy or sell (optional string)

Outputs:
 - The orderbook in a nice little command-line table

If you specify buy or sell you will be given only the buy or sell side of the order book for the specified pair.

## getprice
Getprice will get the price of a pair, based on midpoint of volume of bids and asks

`ocx getprice pair`

Arguments:
 - Asset pair (string)

Outputs:
 - The price / conversion rate of the asset

## placeorder
This will print a description of the order after making it, and prompt the user before actually sending it.

`ocx placeorder name {buy|sell} pair amountHave price`

The price is price, amountHave is the amount of the asset you have. If you're on the selling side, that will be the first asset1 in the asset1/asset2 pair. If you're on the buying side, that will be the second, asset2. 

Arguments:
 - Name (string)
 - buy or sell (string)
 - Asset pair (string)
 - AmountHave (uint)
 - Price (float)

Outputs:
 - Order submitted successfully (or error)
 - An order ID (or error)

## getdepositaddress
Getdepositaddress will return the deposit address that is assigned to the user's account for a certain asset.

`ocx getdepositaddress name asset`

Arguments
 - Name (string)
 - Asset (string)

Outputs:
 - A deposit address for the specified name and asset (or error)

## withdraw
Withdraw will send a withdraw transaction to the blockchain.

`ocx withdrawtoaddress name amount asset recvaddress`

Arguments:
 - Name (string)
 - Amount (uint, satoshis)
 - Asset (string)
 - Receive address (string)

Outputs:
 - Transaction ID (or error)

## getbalance
Getbalance will get your balance

`ocx getbalance name asset`

Arguments:
 - Name (string)
 - Asset (string)

Outputs:
 - Your balance for specified asset (or error)

## getallbalances
Getallbalances will get balances for all of your assets.

`ocx getallbalances name`

Arguments:
 - Name (string)

Outputs:
 - Balances for all of your assets (or error)
