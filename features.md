# Basic building blocks of a "centralized" (custodial) cryptocurrency exchange

 - Wallets
  - Key management

 - Orders
  - Order types, database

 - Matching engine
  - Taking orders from above, keep track of sides, always be matching

 - Authentication

## Notes on type of exchange
There are multiple ways to create an "exchange," one of which is what I'm currently trying to make, but hopefully the APIs I make are robust enough such that the other type can be made as well. There are **Dealer markets** and **Auction markets**. I'm currently making a dealer market, which uses competitive buy and sell orders and a market maker, in this case the exchange, to actually facilitate trades. Auction markets are _very_ similar, in fact the exact same but without a market maker.

Another way of doing exchange, like Arwen is doing, is using RFQ, or **request for quote**. Instead of posting orders (though on the backend it works pretty much the same), you request a quote for what you have, and the exchange gives you a price and offer to accept. We're not using RFQ, we're using limit orders.

## Storage
~~For now we're using Redis but that may change.~~
That changed we're using MySQL because it supports transactions (begin, commit, rollback) and I'd have to implement my own locks for redis that would be annoying, and if it were scaled up then it would be .

I would use a database that is able to be started from go without a system call, but I figure speed is more important for an exchange for now, and if this were go peer-to-peer, there would need to be other changes to the database. Also badgerdb is easy enough to use that it can be switched to super easily if necessary

# Sync
The exchange needs to be synced to determine the number of confirmations a transaction has, and should be if it wants to send transactions.

I also need to figure out a way to remove deposits stuck in orphan blocks on chains not as long as the confirmation variable. This can also be kept along with the block height and transaction. I'm doing this to make sure I can do variable confirmations (based on size of deposit), which should be easy enough to do.
Maybe also keep the block header, I need something to check that it's not an orphan block (check block at height x has block header y). Or there's probably a better way to do this that is just available.

Only once the current height and height stored in the DB differ by the # of confirmations stored along with the deposit, do we make a DB update on the balances table.

# Matching

Here's a rough draft of the matching algorithm for a specific price:
```python
def oppositeSide(order):
    if order.side == "buy":
        return "sell"
    elif order.side == "sell":
        return "buy"
    raise Exception("Order does not have a compatible side")

def IncomingOrderEventHandler(incomingOrder):
    thisSideOrders = append(db.GetSortedOrders(incomingOrder.price, incomingOrder.side), incomingOrder)
    oppositeSideOrders = db.GetSortedOrders(incomingOrder.price, oppositeSide(incomingOrder))

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
 - [x] Signing
   - [x] Sign all messages to/from for identification
 - [x] Matching engine
 - [x] Get all balances
   - [x] Extra command
 - [x] Get Price
 - [x] Remove token stuff in shell
 - [x] ~~Correct dynamic confirmations~~ fixed but I just changed it to 6, but it *could* be made a lot better because I made it easy to do so.
 - [x] **Get robust way of adding multiple tokens**
 - [x] Fix issue with price calculation on sell side
 - [ ] Fix SQL Injection vulnerability lol
 - [ ] Lightning Features
   - [ ] Protocol for custodial only-on-orderbook exchange
   - [x] Lightning fund and push deposit
   - [ ] Lightning withdrawal
   - [ ] Lightning push back on withdrawal channel as deposit
   - [ ] Magic: fully non custodial exchange

## Note on proofs
One thing I've been more and more interested in as a research topic has been the idea of a publicly auditable exchange. This means that things like Provisions will be produced.
One possibility I've been exploring is the idea that liabilities are essentially proof of users for the exchange, and proof of assets (like in Provisions) backs those users up with money. The idea behind this is then the exchange might be able to then "prove" that an action occurred, and that action is backed up by some liabilities.
Provisions then proves that these liabilities are backed up by real assets. 
An exchange can "create" fake liabilities, but those then need to be backed up by assets. 
So the proof of liabilities needs to be reused for both the provisions proof and any other proofs in order for this to work.
This also might be a completely horrible way to do things. Provisions isn't perfect, you need to use P2PK, and it requires users to check their own accounts. It also isn't super efficient, but that might not be avoidable given number of users.
Given the recent SEC report by Bitwise, even considering the fact that the entire thing might not be true at all (I doubt it but you never know), it's clear that wash trading is something that a malicious exchange could do.
One solution to this problem might be a proof that is associated with each order, proving that the order is backed up by assets that the exchange has to be accountable for in the Provisions proof.
I need to think the details through more, like if you should monitor the orderbook, batch order proofs, etc.
I also need to figure out if Provisions is really a good base for this sort of thing.

#### Decentralization notes

One note for decentralization: This is a really good reason why orderbooks shouldn't be kept on blockchains (like some people would want). The occasional 1 or 2 block reorg would screw everything up even if you got everything else solved & right.
We care about the order of orders for matching and matching only, the validity of the orders and settlement can be done elsewhere, like with bitcoin. 
So no reason TO use a blockchain as an orderbook. 
One could make the argument that it's not fair because there isn't time priority, but with blockchains you have transaction fees as priority anyways and all these orderbook-on-blockchain "DEX"es have no issue with using that.
So maybe price/fee priority is the way to go. This at least creates a market for getting an order matched I guess.
Trying to fairly match orders without regarding order they came in is something that would solve a lot of the problems I think. We do want every order to be in the same network / same place though. 
Orderbooks and matching engines exist because we don't know how to do this, maybe the order doesn't really matter, or maybe it does. Figuring out _whether or not_ the order that orders come in matter for fairness of matching is also an interesting problem. 
I've only seen 2 different matching algorithms and they're pretty simple. We can do all the money stuff with bitcoin, so half the battle is done right there. 
We just need to figure out what's a trustless way to store orders all in one place, and how to fairly match them. 
Also, if two people in some sort of DEX network match the same order, only one of them is gonna end up working, and this matters a lot, since if we treated all matched orders as valid, we care which one gets executed since the same buy order could have 2 different sell recipients, and the same sell order could have 2 different buy recipients. This is all assuming there's gossip matching or something super decentralized. It just has so many cases.

#### Notes on arwen

Arwen is actually pretty simple. First you create a user escrow (basically a lightning payment channel) with the exchange, then you tell the exchange what you want and how much they should fund their side. This can be small, this can be large, but the exchange will say yes or no to how much they will fund their channel. This is not the trade, but this is how the exchange shows that they have the funds to trade with you. Then you request for the exchange to give you a quote on the asset in their escrow. If you have bitcoin and told them to open a vertcoin escrow, you could ask them "how many vertcoin can I get for 10BTC?" and they would respond with something like "20000VTC". This can be accepted or denied by you. All of this is pretty simple and it's no different from current exchanges. What's different is settlement. Rather than trading with other people directly, you're trading with the exchange. The actual settlement is a simple atomic swap using HTLC's. 

The exchange needs a lot of funds to work correctly, since someone with a similar amount of money to the exchange can just open up a very large, or very many exchange escrows and user escrows.


### Architecture notes
I've realized that the wallet should probably be decoupled. As the exchange, we just rely on something that takes in transactions and tells us if we've received the money yet, so basically a wallet. We should be able to just connect to that thing, ask it about what addresses it has and has been sent to.

Key management has always been an issue but I'd like to just keep one thing synced up and connect to that, and lit is the most compatible with the chainhook and stuff. I'm sure I'm storing like 3 copies of everything. 

Soo much engineering work that could be done to make this like a robust way of setting up and exchange. Lots of moving parts.
