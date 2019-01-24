# Basic building blocks of a centralized exchange

 - Wallets (user side)
  - keep track of funds / keys

 - Orders
  - Order types, database

 - Matching engine
  - Taking orders from above, keep track of sides, always be matching

 - Accounts

## Wallets

Of course users have their own wallets, but there needs to be some way to make a deposit to a certain account.
I'm guessing the exchange makes an address when you say you want to deposit, or assigns an address to your account. Assigning the address to the account seems like the best idea, because then you can deposit anytime. The private key for account address could be a child private key (BIP32 style maybe) of a private key the exchange owns.

Basically money always comes in to the exchange, the exchange keeps track of your balance (and holds the funds in their wallets so they could pull a mtgox)

## Interacting with the exchange
The exchange needs to have a few functions that the user interacts with:

Un-permissioned commands (and simple mockups of how it might work):

 - Register account
`ocx register username password`

 - Log in (need a way to keep sessions or something)
`ocx login username password`

 - View orderbook
`ocx vieworderbook assetwant assethave`

Or maybe we want dark pools? Could be a feature, probably out of scope, would be difficult to do if you want to match in a decentralized way. Decentralized matching for confidential _orders_ is probably very difficult, aside from the whole decentralized matching problem.

 - Get price (really just getorderbook but with a few more operations)
`ocx getprice`

- Get volume (need to track that server side)
`ocx getvolume`

 - TODO: think of more that you might need

Permissioned commands:

 - Place order
`ocx placeorder price {buy|sell} assetwant assethave amounttobuy`
This will print a description of the order after making it, and prompt the user before actually sending it.

 - Get account's address
`ocx getaddress`
This will return the address that is assigned to the user's account

 - Withdraw
`ocx withdrawtoaddress asset amount recvaddress`
Withdraw will send a transaction to the blockchain.

 - Delete account
`ocx deleteaccount`

For authentication, let's just do some user data storage and send a random token that expires in 30 minutes or something. Server checks token, client stores token and sends it with json.

## Storage
For now we're using Redis but that may change.

I would use a database that is able to be started from go without a system call, but I figure speed is more important for an exchange for now, and if this were go peer-to-peer, there would need to be other changes to the database. Also badgerdb is easy enough to use that it can be  switched to super easily if necessary

# Sync
The exchange needs to be synced to determine the number of confirmations a transaction has, and should be if it wants to send transactions.

This means there needs to be a way of easily interacting with running nodes

# Current features

 - [x] RPC Interface
 - [x] DB Interface
 - [ ] Tesnet interface
 - [ ] Wallets
 - [x] Register
  - [x] RPC Command in interface
  - [x] Database k/v for username and password
  - [x] Respond with generated token
 - [x] Login
  - [x] RPC Command in interface
  - [x] Database k/v for username and password
  - [x] Respond with generated token
 - [ ] Place order
  - [ ] Orders in DB
  - [ ] Different coins in DB
  - [ ] Account balances for said coins in DB
 - [ ] Get balance
  - [ ] Different coins in DB
  - [ ] Account balances for said coins in DB
 - [ ] View orderbook
  - [ ] Orders in DB
 - [ ] Get Address
  - [ ] Testnet interface
  - [ ] Deposit addresses in DB
 - [ ] Deposit
  - [ ] Testnet interface
  - [ ] Deposit confirmations variable
  - [ ] How to confirm deposit
  - [ ] Create master private key
  - [ ] Create derived keys for deposit addresses
