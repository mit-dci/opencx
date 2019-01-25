# cxrpc

This package handles RPC requests coming in to the exchange. Here are all the commands supported so far:
RPC is just a starting point for being able to accept network I/O

The `register` and `login` commands send stuff in plaintext as far as I know so they might be replaced

## UPDATE
register does stuff now but login doesn't. register works the same way and returns nothing

## register
This commands checks that there is no user with the same username, and if there isn't, registers you and returns a token to be included in authenticated commands.

`ocx register username password`

Arguments:
 - Username (string)
 - Password (string)

Returns:
 - Token (byte array)

## login
This commands checks your login credentials and sends you a token, to be included in authenticated commands.

`ocx login username password`

Arguments:
 - Username (string)
 - Password (string)

Returns:
 - Token (byte array)

## getbalance
Once you've registered you can now get your balance:

`ocx getbalance username asset`

This will (unsurprisingly) return your balance if you entered everything correctly.
