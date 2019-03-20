# Basic building blocks of a "centralized" (custodial) cryptocurrency exchange

 - Wallets (user side)
  - keep track of funds / keys

 - Orders
  - Order types, database

 - Matching engine
  - Taking orders from above, keep track of sides, always be matching

 - Accounts

## Important note
There are multiple ways to create an "exchange," one of which is what I'm currently trying to make, but hopefully the APIs I make are robust enough such that the other type can be made as well. There are **Dealer markets** and **Auction markets**. I'm currently making a dealer market, which uses competitive buy and sell orders and a market maker, in this case the exchange, to actually facilitate trades. Auction markets are _very_ similar, in fact the exact same but without a market maker.

Another way of doing exchange, like Arwen is doing, is using RFQ, or **request for quote**. Instead of posting orders (though on the backend it works pretty much the same), you request a quote for what you have, and the exchange gives you a price and offer to accept. We're not using RFQ, we're using limit orders.

## Wallets

Of course users have their own wallets, but there needs to be some way to make a deposit to a certain account.
I'm guessing the exchange makes an address when you say you want to deposit, or assigns an address to your account. Assigning the address to the account seems like the best idea, because then you can deposit anytime. The private key for account address could be a child private key (BIP32 style maybe) of a private key the exchange owns.

Basically money always comes in to the exchange, the exchange keeps track of your balance (and holds the funds in their wallets so they could pull a mtgox)

~~ONCE money comes in to the exchange, send it to a pool address?~~

## Interacting with the exchange
The exchange needs to have a few functions that the user interacts with:

Un-permissioned commands (and simple mockups of how it might work):

 - Register account

`ocx register name`

 - View orderbook

`ocx vieworderbook pair [buy/sell]`

Or maybe we want dark pools? Could be a feature, probably out of scope, would be difficult to do if you want to match in a decentralized way, because of the whole "zero knowledge" thing added on. Decentralized matching for confidential _orders_ is probably very difficult, aside from the whole decentralized matching problem.

 - Get price (really just getorderbook but with a few more operations)

`ocx getprice pair`

This will get the price of a pair, based on midpoint of volume of bids and asks

## NOTE: nothing is permissioned, there used to be authentication but now there's not
Permissioned commands:

 - Place order

`ocx placeorder name {buy|sell} pair amountHave price`

This will print a description of the order after making it, and prompt the user before actually sending it.
The price is price, amountHave is the amount of the asset you have. If you're on the selling side, that will be the first asset1 in the asset1_asset2 pair. If you're on the buying side, that will be the second, asset2. 

 - Get account's deposit address

`ocx getdepositaddress name`

This will return the address that is assigned to the user's account

 - Withdraw

`ocx withdrawtoaddress name amount asset recvaddress`

Withdraw will send a transaction to the blockchain.

- Get Balance

`ocx getbalance name asset`

This will get your balance

- Get all balances

`ocx getallbalances name`

This will get balances for all of your assets.

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
    # or if on the other side SELECT * FROM orders.ltc_btcsellorders WHERE price=x SORT BY timestamp

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

I also need to make some thing in the `match` package to generate the `ltc_btcsellorders` and `btc_ltcbuyorders` table and make sure that orders ALWAYS are the same stuff and you never end up with a unique pair that doesn't exist (because now there would just be the btc/ltc, btc/vtc, ltc/vtc pairs technically, never btc/btc, vtc/vtc, ltc/ltc, vtc/ltc, ltc/btc, vtc/btc) because you only need `n(n-1)/2` pairs to represent all unique pairs. I just have to make sure it's correct and nothing will go wrong.

One thing to think about - Exchanges like Binance have a native asset, BNB, where pretty much every pair goes through. Is it good enough to have a reserve asset like this or should I really add every possible trading pair? It seems like more trading pairs means more possibilities for arbitrage and price manipulation but that guess isn't really based on anything concrete.

# Current features

 - [x] RPC Interface
 - [x] DB Interface
 - [x] Tesnet interface
 - [x] Wallets
   - [x] Withdrawal
   - [x] Test on testnet - debug transaction sending / pushing
 - [x] Register
   - [x] RPC Command in interface
   - [x] Database k/v for username and password
   - [x] Respond with generated token
 - [x] Login
   - [x] RPC Command in interface
   - [x] Database k/v for username and password
   - [x] Respond with generated token
 - [x] Place order
   - [x] Orders and trading pairs in DB
   - [x] Different coins in DB
   - [x] Account balances for said coins in DB
 - [x] Get balance
   - [x] Different coins in DB
   - [x] Account balances for said coins in DB
 - [x] View orderbook
   - [x] Orders in DB
 - [x] Get Address
   - [x] Testnet interface
   - [x] Deposit addresses in DB
 - [x] Deposit
   - [x] Testnet interface
   - [x] Deposit confirmations variable
   - [x] How to confirm deposit
   - [x] Create master private key
   - [x] Create derived keys for deposit addresses
 - [ ] Signing
   - [ ] Sign all messages to/from for identification
 - [x] Matching engine
 - [x] Get all balances
   - [x] Extra command
 - [x] Get Price
 - [x] Remove token stuff in shell
 - [x] ~~Correct dynamic confirmations~~ fixed but I just changed it to 6, but it *could* be made a lot better because I made it easy to do so.
 - [ ] **Get robust way of adding multiple tokens**
 - [x] Fix issue with price calculation on sell side
 - [ ] Fix SQL Injection vulnerability lol
 - [ ] Lightning payment is market order

##
Idea: No registration. The exchange has a single address (or multiple). 
You send to that address, you are depositing. 
Since everything is done through public keys, your public key is extracted from the signatures in the transaction.
Use that public key to sign all orders, etc.
Basically the public key associated with the address you deposited from is now your "account" on the exchange.
You will withdraw to that address as well.
This is also proof of reserves, proof of solvency.

#### Decentralization notes

One note for decentralization: This is a really good reason why orderbooks shouldn't be kept on blockchains (like some people would want). The occasional 1 or 2 block reorg would screw everything up even if you got everything else solved & right.
We care about the order of orders for matching and matching only, the validity of the orders and settlement can be done elsewhere, like with bitcoin. So no reason TO use a blockchain as an orderbook. Trying to fairly match orders without regarding order they came in is something that would solve a lot of the problems I think. We do want every order to be in the same network / same place though. Orderbooks and matching engines exist because we don't know how to do this, maybe the order doesn't really matter, or maybe it does. Figuring out _whether or not_ the order that orders come in matter for fairness of matching is also an interesting problem. I've only seen 2 different matching algorithms and they're pretty simple. We can do all the money stuff with bitcoin, so half the battle is done right there. We just need to figure out what's a trustless way to store orders all in one place, and how to fairly match them. Also, if two people in some sort of DEX network match the same order, only one of them is gonna end up working, and this matters a lot, since if we treated all matched orders as valid, we care which one gets executed since the same buy order could have 2 different sell recipients, and the same sell order could have 2 different buy recipients. This is all assuming there's gossip matching or something super decentralized. It just has so many cases.

#### Notes on arwen

Arwen is actually pretty simple. First you create a user escrow (basically a lightning payment channel) with the exchange, then you tell the exchange what you want and how much they should fund their side. This can be small, this can be large, but the exchange will say yes or no to how much they will fund their channel. This is not the trade, but this is how the exchange shows that they have the funds to trade with you. Then you request for the exchange to give you a quote on the asset in their escrow. If you have bitcoin and told them to open a vertcoin escrow, you could ask them "how many vertcoin can I get for 10BTC?" and they would respond with something like "20000VTC". This can be accepted or denied by you. All of this is pretty simple and it's no different from current exchanges. What's different is settlement. Rather than trading with other people directly, you're trading with the exchange. The actual settlement is a simple atomic swap using HTLC's. 

The exchange needs a lot of funds to work correctly, since someone with a similar amount of money to the exchange can just open up a very large, or very many exchange escrows and user escrows.


### Architecture notes
I've realized that the wallet should probably be decoupled. As the exchange, we just rely on something that takes in transactions and tells us if we've received the money yet, so basically a wallet. We should be able to just connect to that thing, ask it about what addresses it has and has been sent to.

Key management has always been an issue but I'd like to just keep one thing synced up and connect to that, and lit is the most compatible with the chainhook and stuff. I'm sure I'm storing like 3 copies of everything. 

Soo much engineering work that could be done to make this like a robust way of setting up and exchange. Lots of moving parts.
