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

ONCE money comes in to the exchange, send it to a pool address?

## Interacting with the exchange
The exchange needs to have a few functions that the user interacts with:

Un-permissioned commands (and simple mockups of how it might work):

 - Register account

`ocx register username password`

 - Log in (need a way to keep sessions or something)

`ocx login username password`

 - View orderbook

`ocx vieworderbook assetwant assethave`

Or maybe we want dark pools? Could be a feature, probably out of scope, would be difficult to do if you want to match in a decentralized way, because of the whole "zero knowledge" thing added on. Decentralized matching for confidential _orders_ is probably very difficult, aside from the whole decentralized matching problem.

 - Get price (really just getorderbook but with a few more operations)

`ocx getprice`

- Get volume (need to track that server side)

`ocx getvolume`

 - TODO: think of more that you might need

## NOTE: nothing is permissioned, there used to be authentication but now there's not
Permissioned commands:

 - Place order

`ocx placeorder name {buy|sell} pair amountHave price`

This will print a description of the order after making it, and prompt the user before actually sending it.
The price is price, amountHave is the amount of the asset you have. If you're on the selling side, that will be the first asset1 in the asset1_asset2 pair. If you're on the buying side, that will be the second, asset2. 

 - Get account's deposit address

`ocx getdepositaddress`

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
 - [ ] Wallets
   - [x] Withdrawal
   - [ ] Test on testnet - debug transaction sending / pushing
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
 - [ ] Fix issue with price calculation on sell side

#### Decentralization notes

One note for decentralization: This is a really good reason why orderbooks shouldn't be kept on blockchains (like some people would want). The occasional 1 or 2 block reorg would screw everything up even if you got everything else solved & right.
We care about the order of orders for matching and matching only, the validity of the orders and settlement can be done elsewhere, like with bitcoin. So no reason TO use a blockchain as an orderbook. Trying to fairly match orders without regarding order they came in is something that would solve a lot of the problems I think. We do want every order to be in the same network / same place though. Orderbooks and matching engines exist because we don't know how to do this, maybe the order doesn't really matter, or maybe it does. Figuring out _whether or not_ the order that orders come in matter for fairness of matching is also an interesting problem. I've only seen 2 different matching algorithms and they're pretty simple. We can do all the money stuff with bitcoin, so half the battle is done right there. We just need to figure out what's a trustless way to store orders all in one place, and how to fairly match them. Also, if two people in some sort of DEX network match the same order, only one of them is gonna end up working, and this matters a lot, since if we treated all matched orders as valid, we care which one gets executed since the same buy order could have 2 different sell recipients, and the same sell order could have 2 different buy recipients. This is all assuming there's gossip matching or something super decentralized. It just has so many cases.

#### Notes on arwen

Arwen is actually pretty simple. First you create a user escrow (basically a lightning payment channel) with the exchange, then you tell the exchange what you want and how much they should fund their side. This can be small, this can be large, but the exchange will say yes or no to how much they will fund their channel. This is not the trade, but this is how the exchange shows that they have the funds to trade with you. Then you request for the exchange to give you a quote on the asset in their escrow. If you have bitcoin and told them to open a vertcoin escrow, you could ask them "how many vertcoin can I get for 10BTC?" and they would respond with something like "20000VTC". This can be accepted or denied by you. All of this is pretty simple and it's no different from current exchanges. What's different is settlement. Rather than trading with other people directly, you're trading with the exchange. The actual settlement is a simple atomic swap using HTLC's. 

The exchange needs a lot of funds to work correctly, since someone with a similar amount of money to the exchange can just open up a very large, or very many exchange escrows and user escrows.

This would restrict the exchange from opening any more exchange escrows, locking the rest of the users out. This would require a lot of money but if operated in a trustless fashion, would make the exchange completely useless. A competing exchange could pull off this kind of attack, or many arwen-style exchanges could just trade with each other and nobody would be able to trade with the group of exchanges, since they're all locking each other out. The exchanges would probably know that they're all just trading with each other, but if they didn't that would be pretty interesting. The trades *look* like normal people but they aren't. Arwen has a sort of bandwidth based on the funds that it has, due to all settlement happening with the exchange rather than person to person. 


While this definitely gives arwen an advantage since they can be the market maker, you could pull off an attack like this if you don't KYC or have permissioned stuff to make sure 100 medium sized accounts aren't just one gigantic whale trying to sponge out all of your funds into exchange escrows. Arwen tries to solve this with escrow fees, but if you actually did have ~2x the money that the exchange did, you could lock up the exchange for a week. This is why there's still room for improvement for arwen, it can't be completely trustless in order to prevent this kind of attack (correct me if I'm wrong). As far as Arwen is concerned, they should make sure that if they are close to capacity, the fees increase.

If I make an instance of miniArwen, and I only have $5 worth of litecoin, $10 worth of bitcoin, and $8 worth of vertcoin, if Bob creates a channel with like $20 worth of litecoin, and asked for the exchange to collateralize $10 worth of bitcoin, that's all the exchanges bitcoin funds. If you take into account the fact that even with exchange fees, the exchange can't use the fees (since they need to make sure if this is an actual trade they can give rebats) until the user closes the channel, this looks pretty bad for small exchanges. 
Bob with $20 has just ensured, even if they pay tons of money in fees, that my exchange will not be able to trade bitcoin (unless it's with Bob). This still works even if the exchange has a per-address cap, since Bob can just spread his money out accross addresses. 
I think this would be avoided if the fees increased based on how much of the exchanges' funds are not locked in escrows. This is like econimic denial of service. 


However, this depends on how much the exchange owner actually cares about the exchange being up, since my $23 net worth exchange would annoy users who aren't malicious if it costed an exorbitant amount to make a $23 exchange. 
While fees based on reserve capacity could be designed as a deterrent for malicious whales, they also deter users, and stop regular traders from using most of the exchanges' capacity (which is the thing that makes you money after all).
This also doesn't really stop the attacker, since maybe their goal is to stop when the fees get so high that nobody trades on it.

Maybe it's just me, but an exchange that doesn't take custody of your tokens and doesn't match peer to peer seems like a bad business plan because of this. The fee model would have to be really good. It's not that any normal person can screw with this, but the point of bitcoin is that you trust no one, and you especially don't trust your wealthy competitors, even when not in the adversarial bitcoin world.


I mean I'm not sure what you would gain from attacking arwen though, considering it connects to normal exchanges.


Also if I'm wrong then I wrote a lot for no reason? Sucks but it's my fault if I'm wrong so `¯\_(ツ)_/¯`.