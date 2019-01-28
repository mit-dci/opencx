# Basic building blocks of a centralized exchange

 - Wallets (user side)
  - keep track of funds / keys

 - Orders
  - Order types, database

 - Matching engine
  - Taking orders from above, keep track of sides, always be matching

 - Accounts

## Important note
There are multiple ways to create an "exchange," one of which is what I'm currently trying to make, but hopefully the APIs I make are robust enough such that the other type can be made as well. There are **Dealer markets** and **Auction markets**. I'm currently making a dealer market, which uses competitive buy and sell orders and a market maker, in this case the exchange, to actually facilitate trades. Auction markets are _very_ similar, in fact the exact same but without a market maker.

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
~~For now we're using Redis but that may change.~~
That changed we're using MySQL because it supports transactions (begin, commit, rollback) and I'd have to implement my own locks for redis that would be annoying, and if it were scaled up then it would be .

I would use a database that is able to be started from go without a system call, but I figure speed is more important for an exchange for now, and if this were go peer-to-peer, there would need to be other changes to the database. Also badgerdb is easy enough to use that it can be switched to super easily if necessary

# Sync
The exchange needs to be synced to determine the number of confirmations a transaction has, and should be if it wants to send transactions.

This means there needs to be a way of easily interacting with running nodes
That's what the wallet is for

Alright here's the plan:
For every transaction made to one of the wallets, there needs to be a certain number of confirmations that will be tracked. This might be tricky but if it's just in a table with two columns, which are like height and a transaction, then every block it should be pretty easy to make sure that balances get changed. We'll just do a big check every time a block comes in.
I also need to figure out a way to remove deposits stuck in orphan blocks on chains not as long as the confirmation variable. This can also be kept along with the block height and transaction. I'm doing this to make sure I can do variable confirmations (based on size of deposit), which should be easy enough to do.
Maybe also keep the block header, I need something to check that it's not an orphan block (check block at height x has block header y). Or there's probably a better way to do this that is just available.

Only once the current height and height stored in the DB differ by the # of confirmations stored along with the deposit, do we make a DB update on the balances table.

# Matching

Here's a rough draft of the matching algorithm:
```python
def oppositeSide(order):
    if order.side == "buy":
        return "sell"
    elif order.side == "sell":
        return "buy"
    raise Exception("Order does not have a compatible side")

def IncomingOrderEventHandler(incomingOrder):
    # whatever I may have messed up with this whole python thing, this is pseudocode at this point

    thisSideOrders = append(db.GetSortedOrders(incomingOrder.price, incomingOrder.side), incomingOrder)
    oppositeSideOrders = db.GetSortedOrders(incomingOrder.price, oppositeSide(incomingOrder))

    # GetSortedOrders will return orders sorted by timestamp by default
    # This will be something like SELECT * FROM orders.btc_ltcbuyorders WHERE price=x SORT BY timestamp
    # or if on the other side SELECT * FROM orders.ltc_btcbuyorders WHERE price=x SORT BY timestamp

    while len(oppositeSideOrders) > 0 and len(thisSideOrders) > 0
        sideOne = thisSideOrders[0]
        sideTwo = oppositeSideOrders[0]
        if sideOne.volume > sideTwo.volume:
            db.CreditAccount(sideOne.account, sideTwo.volume)
            db.CreditAccount(sideTwo.account, sideTwo.volume)
            db.DeleteOrder(sideTwo)
            thisSideOrders[0].volume -= sideTwo.volume
        elif sideTwo.volume > sideOne.volume:
            db.CreditAccount(sideOne.account, sideOne.volume)
            db.CreditAccount(sideTwo.account, sideOne.volume)
            db.DeleteOrder(sideOne)
            oppositeSideOrders[0].volume -= sideOne.volume
        else:
            volume = sideOne.volume
            db.DeleteOrder(sideOne)
            db.DeleteOrder(sideTwo)
            db.CreditAccount(sideOne.account, equalVolume)
            db.CreditAccount(sideTwo.account, equalVolume)

# I think this algorithm is right

```

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
   - [ ] Orders and trading pairs in DB
   - [x] Different coins in DB
   - [x] Account balances for said coins in DB
 - [x] Get balance
   - [x] Different coins in DB
   - [x] Account balances for said coins in DB
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
 - [ ] Signing
   - [ ] Sign all messages to/from for identification

#### Decentralization notes

One note for decentralization: This is a really good reason why orderbooks shouldn't be kept on blockchains (like some people would want). The occasional 1 or 2 block reorg would screw everything up even if you got everything else solved & right.
We care about the order of orders for matching and matching only, the validity of the orders and settlement can be done elsewhere, like with bitcoin. So no reason TO use a blockchain as an orderbook. Trying to fairly match orders without regarding order they came in is something that would solve a lot of the problems I think. We do want every order to be in the same network / same place though. Orderbooks and matching engines exist because we don't know how to do this, maybe the order doesn't really matter, or maybe it does. Figuring out _whether or not_ the order that orders come in matter for fairness of matching is also an interesting problem. I've only seen 2 different matching algorithms and they're pretty simple. We can do all the money stuff with bitcoin, so half the battle is done right there. We just need to figure out what's a trustless way to store orders all in one place, and how to fairly match them. Also, if two people in some sort of DEX network match the same order, only one of them is gonna end up working, and this matters a lot, since if we treated all matched orders as valid, we care which one gets executed since the same buy order could have 2 different sell recipients, and the same sell order could have 2 different buy recipients. This is all assuming there's gossip matching or something super decentralized. It just has so many cases.
