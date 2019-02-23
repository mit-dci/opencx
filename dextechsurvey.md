# Decentralized exchange tech survey

I'm going to try to review as much as possible about decentralized exchange technology

## Research, proposed DEXes, anything not currently implemented
It's good to get a survey of the research and not-yet-implemented solutions to problems that people supposedly have with exchange. Some things here are DEXes, some are techniques used to maybe decrease trust needed in exchange, but most are not formal and not peer reviewed.
 - [x] Arwen
 - [x] Nash / NEX
   - [x] [NEX Whitepaper](assets/whitepaper_v2.pdf)
   - [x] [The heart of Nash: Our off-chain matching engine](https://medium.com/nashsocial/the-heart-of-nash-our-off-chain-matching-engine-499cf2c23670)
 - [x] [gnosis/dex-research](https://github.com/gnosis/dex-research)
   - [x] [dFusion](https://github.com/gnosis/dex-research/tree/master/dFusion)
   - [x] [Batch Auction Optimization](https://github.com/gnosis/dex-research/tree/master/BatchAuctionOptimization)
   - [x] [Plasma Research](https://github.com/gnosis/dex-research/tree/master/PlasmaResearch)
 - [x] Vitalik Reddit: [Let's run on-chain decentralized exchanges the way we run prediction markets](https://www.reddit.com/r/ethereum/comments/55m04x/lets_run_onchain_decentralized_exchanges_the_way/) 
 - [x] ethresear.ch ["Decentralized exchanges" category](https://ethresear.ch/c/decentralized-exchanges) as of 2/19/2019
   - [x] [Introducing DTRADE. Confoederation dapp](https://ethresear.ch/t/introducing-dtrade-confoederatio-dapp/4870)
   - [x] [Self-referential proxy exchange - does this exist?](https://ethresear.ch/t/self-referential-proxy-exchange-does-this-exist/4515)
   - [x] [Batch auctions with uniform clearing price on plasma](https://ethresear.ch/t/batch-auctions-with-uniform-clearing-price-on-plasma/2554)
   - [x] [DutchX - fully decentralized auction based exchange](https://ethresear.ch/t/dutchx-fully-decentralized-auction-based-exchange/2443)
   - [x] [A note for the Dutch Team & other Dapp developers](https://ethresear.ch/t/a-note-for-the-dutch-team-other-dapp-developers/2637)
   - [x] [Improving front-running resistance of `x*y=k` market makers](https://ethresear.ch/t/improving-front-running-resistance-of-x-y-k-market-makers/1281)
   - [x] [Limit orders and slippage resistance in `x*y=k` market makers](https://ethresear.ch/t/limit-orders-and-slippage-resistance-in-x-y-k-market-makers/2071)
   - [x] [Reverse Parimutuel Options on Bitcoin](https://ethresear.ch/t/reverse-parimutuel-options-on-bitcoin/1816)
 - [x] Binance Chain

## Topics
These topics are for me to review and asses what the benefits of each are and how they would be relevant in an exchange. Most of these are reviewed topics as well as used for more general things.
 - [x] "Provably fair" matching
   - [x] Difference between verifiable computation and zero knowledge proofs
 - [x] Proof of assets
 - [x] Cross chain swaps

## Papers about crypto exchanges specifically
 - [ ] [Decentralized Cryptocurrency Exchange Security Analysis 6.857 Project](https://courses.csail.mit.edu/6.857/2018/project/Hao-Chang-Lu-Zhang-CCExch.pdf)
   - Written for 6.857 by Parker Hao, Vincent Chang, Shao Lu, and Chenxing Zhang
   - All CSAIL except Shao Lu, who is Harvard.
 - [ ] [Tesseract - Real Time Cryptocurrency Exchange Using Trusted Hardware](https://eprint.iacr.org/2017/1153.pdf)
   - IC3 Preprint by Cornell, Cornell Tech, SJTU, and Eth Zürich
 - [ ] [Cryptographic Securities Exchanges](http://www.eecs.harvard.edu/~cat/cm.pdf)
   - By Christopher Thorpe and David C. Parkes, from EECS at harvard.
 - [ ] [Deconstructing Decentralized Exchanges](https://stanford-jblp.pubpub.org/pub/deconstructing-dex)
   - Essay by Lindsay X. Lin from Interstellar, published in the Stanford Journal of Blockchain Law and Policy
 - [x] [Provisions: Privacy-preserving Proofs of Solvency for Bitcoin Exchanges](https://users.encs.concordia.ca/~clark/papers/2015_ccs.pdf)


## Centralized exchanges
It's also good to see what centralized exchanges could do if they were malicious - exploring attack vectors on the user and exchange.
 - [x] OpenCX (lol)
 - [x] Coinbase
 - [x] Binance
 - [x] MtGox
 - [x] Bitfinex
 - [x] Smaller exchanges

## Implemented DEXes and exchange problem solutions
These are currently implemented "solutions" to problems that users supposedly have with cryptocurrency exchanges. I'll determine whether or not these actually solve any problems, and how well they solve them. One thing that I will be covering a lot is whether or not they are platforms which are bound to a single cryptocurrency, and rely on the fact that said currency is the biggest in order to solve what they set out to solve.
 - [ ] Komodo
 - [x] BitShares
 - [x] 0x
 - [x] Kyber Network
 - [x] EtherDelta
 - [x] IDEX
 - [x] Uniswap

# Research, proposed DEXes, anything not currently implemented

## Arwen
So this is basically going to be my summary and review of Arwen as a cross-chain non-custodial exchange.

Arwen starts out by exploring the current DEX landscape, explaining that p2p exchanges don't have liquidity if they don't have users, which is true. They also point out that at the current moment, centralized exchanges have the most liquidity currently.

This is one of the reasons why arwen chooses to interact with centralized exchanges. They then summarize ethereum DEX protocols and how their atomic swaps use smart contracts but only really work with ETH and ERC20 tokens.


Arwen then summarizes ther "TierNolan" protocol, which is really the HTLC's we know and love.

Arwen then discusses the issues with settling HTLC's and broadcasting trades on-chain:
 - Speed: Block times are long so you have to wait a long time to do stuff with your orders.
 - Scalability: Healthy trading means lots of trades, lots of trades means lots of transactions, which we currently can't handle, and is hard to scale on chain.
 - Front-running: Broadcasting your trades on chain leads to race conditions.

The Arwen setup has two "escrows," which are extremely similar to lightning channels, although they're defined to only be between the user and the exchange.

These channels are opened by publishing a funding transaction to the blockchain. They're multisig channels, 2 of 2 (Between the user and the exchange), and allow the same atomic swap and HTLC functionality as lightning. However, Arwen needs to withstand transaction malleability attacks since it wants to support non-segwit chains.

Arwen then goes through different types of lockup griefing and their solutions, since they use HTLC's and HTLC's are vulnerable to lockup griefing. They point out that lockup griefing is very plausible in lightning due to potentially lots of intermediary nodes.

They claim that as the exchange, there is no incentive to do lockup griefing since they make their money through trading, and that would destroy the exchanges' reputation, making the exchange earn less money.

The user interacting with the exchange, however, could want to do lockup griefing. Arwen avoids this by making the user lock up their coins first, but also requires the user to pay a fee for the time that they are locking up the exchanges' assets. The user then has incentive to trade and close the exchange escrow when they are done. The exchange does give a rebate for time not used.

They make the point that the only one who has the right (not the obligation) to execute the trade is the one that picked the preimage of the hash or "puzzle" in the HTLC, which is why in Arwen, the exchange is the one who does so.


Arwen uses RFQ, which is basically this:
 - User: "I'd like some BCH for my 1 BTC"
 - Exchange: "Yeah you can get 10 BCH for 1 BTC for the next 10 minutes"
Then the exchange and user do an atomic swap using an off chain HTLC, much like lightning.


Limit orders work the same way, but instead the user just proposes what they want and what they have, and the exchange just doesn't execute the trade (and reveal the preimage to the hash in the HTLC) until that's the market price. The user can also always cancel these by closing the escrows.

There's a note about unidirectional channels and how publishing the most recent transaction in arwen is always beneficial.

Finally, there are a lot of diagrams about their implementation of limit orders and RFQs, and they also explain bidirectional RFQ's, where if you bought one asset then you should be able to also sell it back with another RFQ. It's a very dry paper but it's a very solid idea. It is still a market maker but as far as non custodial exchanges go, this is a good way to do it.

Also in the arwen paper there's a mistake, on the last page it says that "Plasma [28] is a proposal for a layer-two decentralized exchange protocol on Ethereum", which is blatantly false. Plasma is a proposal for layer-two off-chain transactions on ethereum, much like lightning but for ethereum. 

**Opinion warning:** I don't think this paper should have been this long. It overcomplicates extremely simple ideas and the idea of arwen, even the "advanced" stuff like limit orders and bidirectional RFQ, is super simple. It seems like lightning but you know the person on the other side is an exchange, and the exchange assumes you're a client that wants to use the exchange, so you have a protocol for talking to each other. Compared to the level of ideas that have been discussed on the #bitcoin-wizards IRC, this definitely didn't need to be 21 pages.

## Nash / NEX

### NEX Whitepaper Review
Their whitepaper is now off of their site, but luckily I had it downloaded so it's linked in the checklist.

NEX does a good job in summing up the Decentralized Exchange space, tradeoffs, and earlier approaches. They point out that decentralized exchanges that place orderbooks directly on the blockchain run price-time priority matching redundantly across all machines on the network, and run very slowly as a result. Here are some of the points about decentralized exchange taken from the paper:
 - On-chain orderbooks and matching contracts or matching validation logic is very slow and very redundant.
 - Automated market maker smart contracts are also very slow and could lost money based on the amount of liquidity they provide
 - Centralized matching is the most efficient option
 - Cross-chain exchange is difficult to do since traders must find counterparties first
 - Cross chain relays such as kyber network cannot operate on bitcoin, or any other cryptocurrency without turing-complete smart contracts, and requires a trusted oracle

They then lay out their architecture, which is essentially a trusted matching engine and smart contracts on multiple chains. The matching engine is supposed to be deterministic and publicly verifiable. The cross chain exchange is done by the trusted matching engine, using the smart contracts on each chain, and orders must be signed by the user for orders to be placed. This is done through the smart contract. Withdrawal also happens with the smart contract, and that must be signed as well. These signatures and smart contracts are supposed to prevent the matching engine from making orders or withdrawing on the user's behalf, or doing anything that would decrease the balance of a user.

Here's a basic run-through of what would happen:
 1. User calls a smart contract function to deposit money.
   - Anything that interacts with that balance must be a signed, authenticated action.
 2. User calls a smart contract function to place an order, and that action must be signed. If successful, the balance goes to a matching engine escrow type place, the order is placed and the matching engine records the fact that it was (for user verification).
   - If the user would like to cancel their order that is also a smart contract function, and must be signed. The act of placing an order essentially puts your funds in an escrow that you can pull out of _if you cancel your order / decide to withdraw_ or that the matching engine can use to settle trades.
 3. The matching engine matches an order and transfers the money, using smart contracts on either chain. The matching engine is still an authority, but non-custodial as far as your deposits in the contract.


This is one of the most thoughtful pieces of decentralized exchange research, since they fully recognize that the matching engine will have to be trusted, and count it as a trade-off. The efforts towards making the matching less centralized are also thoughtful.

### The Heart of Nash Article review and thoughts
This article is essentially a tip-of-the-iceberg explanation of the Nash matching engine, which claims to be a "provably fair, distributed system". This is the selling point of their decentralized exchange technology. They openly say that they are trading off trust in matching for the huge speed bonus you get by not having extremely redundant computation, like other DEXes that have their matching not off chain. They acknowledge this by requiring that their matching engine be provably fair and deterministic, so users can verify there is nothing bad happening. 

Their matching engine is a distributed system with its own consensus algorithm.

They explain that users use smart contracts to sign and place orders, and their matching engine supposedly also has a "heartbeat", which is something used to ensure determinism between parallel market processes. They don't really elaborate on the technical details at all.

## Gnosis DEX Research
This is a repo by gnosis called dex-research so I decided to check it out.

It's an interesting look into batching orders as a form of auctioning an asset against another asset that uses zero knowledge proofs and plasma stuff. As a decentralized exchange, it's an interesting protocol but it's also yet another ethereum smart contract "DEX".

### dFusion
dFusion is sort of a smart contract that will take in a bunch of orders, freeze them, then have people propose which orders to match using zk-SNARKS. Not implemented yet. [Here's a link to the ethresear.ch post](https://ethresear.ch/t/building-a-decentralized-exchange-using-snarks/3928), where they say that each snark will cost $1200 to produce on AWS. The incentive to spend $1200 would be that you get trading fees from the trades. They try to implement their batch auction optimization too, which makes since, it's literally one giant batch trade. They say auction closing / matching times would be anywhere from 3 hours to 1 day. The matching optimizes for "trader surplus" or "trading volume". The high cost of calculating the SNARK is a "crypto-economic" incentive to only post valid solutions. They also collect orders on chain, although they would probably do it on a plasma chain, so it's still ethereum only.

### Batch Auction Optimization
This is basically a way of making sure low liquidity tokens are sorta liquid in their system of plasma batch auctions.


### Plasma Research
This part is just TeX but it defines how you would do batch auction stuff on plasma. Again, not really something that would be interesting for someone making a decentralized exchange that isn't just an erc20 token swap contract on ethereum.

## Vitalik reddit
In this reddit post he proposes an automated market maker smart contract, and suggests those be used for decentralized exchange contracts on ethereum. I think this is where he proposed the idea for uniswap, since one of the two in the github organization for uniswap referenced it.

## ethresear.ch DEX topics
The ethresear.ch DEX category seems to be pretty weak when it comes to talking about actual decentralized exchange, mostly just proposals for how to scale a DEX that is only on ethereum, or posts about a "new DEX Dapp". Now for the reviews:

### DTRADE / Confoederation Dapp
This post is really just a link to a blog post about a platform called "Confoederation TRADE," which, from reading the actual post as well as the website, seems to be a collection of smart contracts on ethereum that are neither implemented nor justified. Reading the blog post and the website was a waste of time.

### Self-referential proxy exchange - does this exist?
This post discusses an idea the original poster calls a "self-referential proxy exchange". Here are the requirements, as quoted from the post:
> 1. The exchange has its own token.
> 2. The exchange can accept and hold various crypto assets in exchange for its own token.
> 3. The holders of that token can vote on the exchange rate between the token and other crypto assets (individually).

Supposedly the point is that the users of the system are participating in a sort of gigantic futures contract. Not really relevant to decentralized exchange, but could be relevant to maybe stablecoins or implementable as a smart contract.

### Batch auctions with uniform clearing price on plasma
This is a post by one of the gnosis people trying to spur conversation about their papers on batch auctions for eliminating front-running and bundling liquidity. Here is the basic idea:

Consider the ethereum root / plasma chain model. The plasma chain is necessary for throughput as well as slashing conditions for collusion. The first step is that a set of "bonded" participants participate in distributed key generation. If they try to reveal secret messages that were exchanged in the distributed key generation process before the auction is over, they are subject to slashing. Then, those who would like to place orders will place them within a set period of time (the auction time), and once that time period is over the orders are revealed, price is calculated, and proofs for order settlement / matching are generated. 

This is another one of those things that seems only useful on ethereum and doesn't really have much to do with decentralized exchange, other than their $1200 snark-based order matching, which is also essentially their version of a proof of work but it's on ethereum so _slashing fixes everything_.

### DutchX - fully decentralized auction based exchange
This post is yet another on-chain ethereum DEX Dapp with one difference, it uses many dutch auction (2 per token pair) smart contracts. They are just contracts so any ERC20 token should be able to be listed, and there isn't much new being done.

I'm not going to explain much in detail about the plethora of ethereum-or-otherwise based smart contract DEXes, because they usually have orders on-chain, do matching on-chain, are extremely slow, and increase the size of already large ledgers. There's tons of redundant computation being done that doesn't provide security guarantees, only fairness in the sense that the price-time priority matching algorithm is fair based on the timestamps of the order when the timestamping is being done on the blockchain.

DutchX is also another Gnosis thing. They seem to like auction based exchange.

### A note for the Dutch Team & other Dapp developers
This is just a complaint that DutchX isn't available in certain countries.

### Improving front-running resistance of `x*y=k` market makers
Vitalik here addresses the issues of miners being able to front-run automated market maker based exchanges for profit. He proposes a way to turn that profit into the very least a griefing attack by the miner.

### Limit orders and slippage resistance in `x*y=k` market makers
The problem discussed in this post is basically that once you start taking up a large amount of the liquidity of the automated market maker, there's a high rate per token. The OP of this post is one of the two uniswap members, and he proposes a way to use on-chain limit orders as well as the automated `x*y=k` market maker to fill orders with high volume as well as those with low volume without too much "price slippage" due to trade volume increasing rate, in the nature of `x*y=k` market makers. 
The limit orders are not necessarily executed very quickly but they are executed without affecting the issues of placing high volume trades on automated market makers.

### Reverse Parimutuel Options on Bitcoin
This post basically proposes a way of trading using standard option calls, puts, and derivatives marktes to have a sort of "insurance" for companies and miners who need to protect from a falling market. It's essentially leverage, nothing about DEXes here, not really sure why it's in the DEX category

### Conclusion on the ethresear.ch DEX category
As expected, the only discussion worthwhile here is for those who would like to discuss strategies for doing advanced types of decentralized exchange for ERC20 tokens. The stuff about automated market makers is cool, but the rest really isn't relevant.

## Binance Chain
Binance chain is basically bitshares with a different consensus algorithm. They use tons of tendermint software. Orders are on chain, matching is part of validation. There are one second block times.

# Topics

## "Provably fair" matching
Provably fair matching is mostly a topic brought up by NEX / Nash. The idea is to make sure that the matching is completely deterministic so that anyone on any computer can verify that each step is being done correctly. I'm not sure if anyone will verify it, or if anyone will be able to, considering the whole point of an off chain matching engine is to remove computing redundantly, but it's an interesting idea. 

The actual question of "fairness" seems like a topic that is up to interpretation. Many believe that the price-time priority matching algorithm is the only fair algorithm. I'm not sure whether this is the case. 
With price-time priority, there are issues with front-running and race conditions. There would be similar issues with bitcoin transactions if it didn't use transaction fees as incentive for miners to pick up transactions. With large distributed networks it would be difficult to determine "true" order of transactions anyways, since there would be conflicting opinions on which came first depending on how the transaction were propagated. 

Maybe price-fee priority makes the problem of decentralized matching much simpler, by removing a constraint that was put there in order to enforce undefined "fairness" (time).

One other issue with "provable fairness" is the issue that if you were to submit a log of all inputs, the current state, and a program to allow the user to verify that the inputs and program lead to the current state, then the user would take an equal amount of time to verify that the computation happened correctly. This isn't exactly feasible for most, since if we were to do this for Google then it's not like we all have servers that match up to Google's. 

Also this means if you did this for a program that doesn't stop, you essentially have to be running a clone of the system that the prover is running in order to keep up. So it would be better if we could verify that an exchange did some algorithm correctly in less time than it took the algorithm to run, hopefully constant or something with better asymptotic performance.

## Proof of assets
Proof of assets and proof of solvency is something that has not been implemented in any major cryptocurrency exchange to date. The research for this topic was first introduced by Gregory Maxwell and Peter Todd in the bitcoin-wizards IRC. The Provisions paper suggests using additively homomorphic pederson commitments for zero knowledge proofs for proofs of solvency with certain properties.

The provisions paper points out that "proof of reserves without a corresponding proof of liabilites is not sufficient to prove solvency."

This means that it doesn't matter whether or not you can prove you have a certain amount of money to prove solvency unless you can prove you owe an equal or lesser amount of money. If you owe more money than you have, then you're not completely solvent.

The difference between a proof of reserves and proof of assets is that a proof of reserves prove ownership of assets reserved to cover certain liabilities, where a proof of assets is just proof of ownership of assets.

Unfortunately Provisions only works on pay to pubkey addresses, and can't prove ownership of P2PKH addresses or P2SH addresses. They acknowledge this limitation in the paper, and say it is "an interesting challenge for future work."

## Cross chain swaps
For those who would like to create exchanges, either centralized or decentralized, one feature that is very important to users is the ability to exchange two different cryptocurrencies. This seems obvious, but current "decentralized exchanges" can only exchange token on one ledger, or one chain.
Arwen may be the only "exchange" that has atomic swaps across two different ledgers / chains. This, along with advanced interoperability in general, is a topic that is not very well researched but is a very important area of research.

Some, like Nash, choose to settle trades by having the matching engine simultaneously call smart contracts, but one side of settlement doesn't actually guarantee the other side of settlement. This is why audits and provable fairness are necessary.

# Papers about crypto exchanges specifically

## Decentralized Cryptocurrency Exchange Security Analysis 6.857 Project

## Tesseract - Real Time Cryptocurrency Exchange Using Trusted Hardware

## Cryptographic Securities Exchanges

## Deconstructing Decentralized Exchanges
TODO: summary of essay and general opinion

The essay starts out by defining the architecture of a decentralized exchange, however it doesn't get one thing completely right. They state,
> A decentralized exchange application builds on top of a decentralized exchange protocol, and adds an on-chain or off-chain order book database and a graphic user interface (GUI) and/or APIs so that the information is easily accessible.

This isn't necessarily true, as some decentralized exchange applications, like automated market makers, have no use for an order book. One example of this is uniswap, which is an `x*y=k` market maker.

But when it comes to their listing of the components of a decentralized exchange architecture, they are spot on:
> 1. The blockchain platform & technical implementation
> 2. The counterparty discovery mechanism
> 3. The order matching algorithm
> 4. The transaction settlement protocol

## Provisions: Privacy-preserving Proofs of Solvency for Bitcoin Exchanges
Provisions starts out by describing the idea of proofs of solvency. As stated somewhere else in this document, a proof of reserves (or assets) is not sufficient without a proof of liabilities. The paper recalls how the idea was introduced by Gregory Maxwell, and describes his solution to the problem.
Gregory Maxwell's solution used a merkle tree, summing the balances of the leaf nodes in the parent nodes, and including the sum of the child balances when hashing and concatenating the child hashes. So the actual hash of the children would be `h(Sum,leftHash,rightHash)`. However, when the exchange is proving that the user's account is included in its liabilities, it also reveals the sibling node, so it reveals the liabilities for the account in the sibling node.
This also exposes the total liabilities of the exchange. The more proofs are made, the more information is revealed about other accounts.

Provisions proposes a zero knowledge proof based solution to proof of solvency. One issue with doing this is that multiple exchanges could possibly collude to create a valid proof of solvency.
However, they provide an extension to provisions which is able to prove that one exchange is not sharing addresses with another exchange running Provisions by also providing the result of a deterministic function on the private key. This does reveal the number of addresses (or private keys) the exchange controls though.
There are three protocols in Provisions: 
 1. Proof of assets
 2. Proof of liabilities
 3. Proof of solvency

Cryptographic primitives in Provisions:
 1. The secp256k1 curve is used as the group G, with fixed public generators g and h. G is of prime order q.
 2. Pederson commitments
   - The commitment to a message m in Z_q is defined as `com=(g^m) * (h^r)`, where r is chosen at random in Z_q.
   - g is defined as being the standard g from secp256k1.
   - h is defined as the hash of the string `Provisions`
 3. Non-interactive Zero-Knowledge Proofs (NIZKP)
   - The paper says that they can be adapted from basic sigma protocols (like schnorr proof of knowledge of discrete log).
   - Any alternative sigma protocol to NIZKP compilation is sufficient (so just fiat-shamir whatever sigma protocol with knowledge soundness).
   - Easiest proof of discrete log for me is the fiat-shamir'ed schnorr: 
     - `s = k - h(R,m)a <=> sG = kG - h(R,m)aG <=> sG = R - h(R,m)A`

Assumptions in Provisions:
 1. The bitcoin blockchain is available for all parties to compute the quantity of bitcoin owned by each address.
   - In the paper, they define a y in the group G as a public key, and use `bal(y)` to denote the balance associated with y.

First they describe **proof of assets**.

The exchange selects a set of Bitcoin public keys, denoted PK as `y_1,...,y_n`, and describes `x_1,...,x_n in Z_q` as the secret keys so that `y_i = g^(x_i)` for i from 1 to n. They then define S to be a subset of PK for which the exchange knows the private key. Each `s_i` is a boolean, either 0 or 1, which indicates whether or not the exchange knows the secret key corresponding to that `y_i`.
They then define `b_i` to be `g^(bal(y_i))`.

They define `Assets` to be equal to the summation, with i from 1 to n, of `s_i * bal(y_i)`

The individual pederson commitments for `bal(y_i)` (which remember, can be added homomorphically!) are defined as:

`p_i = h^(v_i) * (b_i)^(s_i)`, where `v_i` is picked randomly in `Z_q`.

So we can define the commitment `Z_Assets` as being the product of all `p_i` from 1 to n, yielding `Z_Assets = h^(sum of all v_i) * g^(Assets)`

They then define other auxiliary values (not going to go through the rest of the math but it checks out and is sorta cool) to create a sigma protocol, an honest-verifier zero knowledge proof (HVZKP) for privacy-preserving proof of assets.

Then they describe the **proof of liabilities**.

Proof of liabilities again uses pederson commitments (or zk-SNARKS, depending on which version of the paper you read) with a couple changes so that the user can verify that their balance is included in the commitment to total liabilities on their own. Public auditors could also verify that each of the commitments to balances add to the commitment of total liabilities. This is another HVZKP.

Finally they describe the very simple **Proof of solvency**.

1. Exchange runs the first protocol to generate `Z_Assets`.
2. Exchange runs the second protocol to generate `Z_Liabilities` and the list of liabilities.
3. Exchange computes `Z_Assets * (Z_Liabilities)^(-1) = Z_(Assets-Liabilities)`
4. Exchange proves in zero knowledge that `Z_(Assets-Liabilities)` is a commitment to the value 0.

They then discuss a variation for fractional-reserve exchanges, and describe the proof of non-collision, proving that exchanges are not covering each other's liabilities.

They do describe a situation where an exchange excludes a certain set of user balances, where they're not that sure that those users will check that the proof of solvency includes their balance. They calculate the probability, given that a random set of accounts check.

Provisions also does not provide dispute resolution, and "If a user finds their account missing or balance incorrect, they do not have sufficient cryptographic evidence that this is the case". They say that it "appears unsolvable cryptographically". It's also not possible to know whether or not a user who says their balance isn't included, is correct.

The goal for efficiency is to be able to produce a privacy-preserving proof of solvency for Coinbase, with roughly 2 million users.

P2PKH and P2SH addresses can also not be used, only public keys that are on the blockchain.

# Centralized Exchanges

## OpenCX (lol)
This is my small SQL-injection vulnerable exchange that gets about 100 transactions per second. There's a ton of stuff that could be done to increase performance, but it's completely custodial for now and is a good prototype exchange similar to ones out there today.

## Coinbase
Coinbase is massively popular and holds a very large percentage of many cryptocurrencies. It has been audited by Andreas antonopoulos, and is trusted by many users around the world. However, it's also completely custodial and, if it wanted to, could steal all of it's users cryptocurrency. Being one of the biggest exchanges, it could very well legitimately liquidate all of its users assets through its own platform.

Coinbase was also hit with a DoS attack when it initially said it wouldn't support Bitcoin Cash.

## Binance
Binance is a very large, very popular crypto exchange that also is well trusted

## MtGox
MtGox is infamous for being the largest and one of the most trusted cryptocurrency exchanges which had then "lost" approximately 850,000 bitcoins. It's suspected there had been bitcoin continously stolen out of its hot wallets since 2011, 3 years before the bitcoins were reported as being "lost". Mt Gox also participated in its own proof of reserves by moving 424242 bitcoins from cold storage to a new address, after a hacker dropped the price of bitcoin to one cent. There were various security issues with Mt Gox, and bitcoin was far from mature or perfect at that point.

## Bitfinex
Bitfinex was also previously the largest cryptocurrency exchange that also "lost" cryptocurrencies, about 120000 bitcoins were stolen from bitfinex.

## Smaller exchanges
Bitstamp has participated in its own "proof of reserves" by moving all of its assets to a new address.

# Implemented DEXes and exchange problem solutions

## Komodo
I'm going to try to get through this 99-page whitepaper in an effort to figure out if there are any strengths or weaknesses to this protocol.

Right off the bat, Komodo claims to be leaders in the field with their atomic swap technology, and apparently have some privacy features as well.

There are 5 parts to the Komodo Whitepaper:
1. Komodo’s Method of Security: Delayed Proof of Work (dPoW)
2. The Decentralized Initial Coin Offering
3. Komodo's Atomic-Swap Powered, Decentralized Exchange: BarterDEX
4. Komodo's Native Privacy Feature: Jumbler
5. Additional Information Regarding the Komodo Ecosystem

### Komodo's Method of Security: Delayed Proof o Work (dPoW)
This part has lots of explanation about the foundations of consensus protocols, and in particular why Proof of Work is valuable.
They first go on for a couple pages explaining encryption and proof of work in layman's terms, and, in the paper's own words:
> The following descriptions are simplified explanations of a truly complex byzantine process. There are many other strategies cryptocurrency miners devise to out-mine their competition, and those strategies can vary widely.
This is all part of a 6-page subsection, "What is a Consensus Mechanism?"
They then explain other aspects of proof of work including environmental effects, 51% attacks, and the longest chain rule.
They explain Proof of Stake, and compare it to Proof of Work.

18 pages in, they start to explain Delayed Proof of Work.
dPoW does not use the Chain with the most work for all blocks. Komodo has a stake-weighted vote to elect 64 separate notary nodes who notarize blocks.
The system requires notary nodes to generate a hash of a block hash on the Komodo network, the block height of that block on the Komodo network, and the letters "KMD".
This gets published to the Bitcoin blockchain (or any other with Proof of Work) using an `OP_RETURN`. 
This hash will get concatenated with the txid on whatever chain it's published on, and then that will be submitted with a notarization transaction to the Komodo chain.

Every 65 blocks, a notary node gets the chance to mine the Komodo chain on "easy mode" (which I think is 10 zeros at the beginning of a block). Each notary is on its own cycle, so for 64 blocks notary nodes will be the only miners. Non notary nodes can theoretically mine during this period but probably won't because easy mode is so easy.

Every 2000 blocks, Komodo removes the ability for notary nodes to mine on easy difficulty for 64 blocks. After this period, notaries keep mining.

## BitShares
BitShares keeps an orderbook on-chain, and the matching algorithm is also a part of the validation logic. You can issue your own assets, and create a whole bunch of fancy "SmartCoins" and Collateralized tokens. There's also some form of margin trading.

They use "Delegated Proof of Stake" consensus.

## 0x
0x is basically EtherDelta (see below) but with a distributed / decentralized orderbook. Makers interact with "relayers" who maintain the order book. This replaces the EtherDelta servers' orderbook in the EtherDelta example. Relayers also take fees.

## Kyber Network
I'm not 100% sure how the kyber network works, but I know it has a bunch of reserves and reserve operators for different tokens. The reserve operators are in charge of setting the exchange rates so the reserves don't run out?

As always, everyone interacts with a smart contract and users will convert their tokens at a market rate with this smart contract.

It seems somewhat similar to uniswap but not automated and third parties can provide reserves of any ERC20 token.

There's also a "network operator" that determines which tokens are listed and delisted.

## EtherDelta
[So the EtherDelta Founder was charged by the SEC for operating an unregistered exchange. He's paying a total of $388,000 in fines.](https://www.sec.gov/news/press-release/2018-258)

EtherDelta keeps its orderbook off chain, and it also uses a smart contract for funds management. Someone who wants to place an order will communicate with an EtherDelta server with a signed order, and the EtherDelta orderbook will be updated. Takers look at the orderbook, and use orders as input to the smart contract with another matching order.

Then the smart contract verifies that the signature originated from the order maker, and that the order is still valid. When that's verified, funds are transferred from one user to another in the smart contract, and users can withdraw. It does only work with ethereum tokens but at least it works with ethereum tokens.

## IDEX
IDEX uses smart contracts for settlement but stores balances, the orderbook, transaction queue, and matching engine on IDEX servers. The user deposits into a smart contract and places an order using that smart contract. The IDEX server is authenticated to execute signed orders on the smart contract, and when it matches orders it calls a smart contract method to settle two smart contracts. Balances are updated and users can withdraw if they'd like.

## Uniswap
Uniswap is an implementation of an `x*y=k` market maker, where there is some constant `k`, and `x` is an amount of one asset stored in the smart contract, and `y` is an amount of another asset stored in the smart contract. Suppose I created a `x*y=k` market maker, specified the constant to be `100`, and funded it with 10 BTC as `x` and 10 LTC as `y`. 

The price is 1 BTC/LTC.

Now I want to buy BTC. If I want to buy 1BTC then I would have to provide the amount of LTC which would keep `k` at 100.

So `y = k/x`, `x=9` so I would have to provide `100 - 100/9`, or `1.11111111` LTC to the smart contract. This would make the price `x/y` or `y/x`, depending on which side of the pair you're trying to get the price of. It's really simple and is another way of looking at making markets.

# Notes
**Orders** specify a transfer of two assets. That's it. It essentially specifies 2 desired transactions.

**Orderbooks** are sets of orders. Not all exchanges need orderbooks.

**Matching algorithms** are required to decide which orders get executed, and when.

`x*y=k` is technically a matching algorithm.
Price-time priority is also a matching algorithm.
Pro-rata is also a matching algorithm.

**Settlement** is the term for executing the transfer of assets between two or more accounts / assets.

Most settlement is custodial, since non-custodial matching and settlement is sorta difficult. There's a reason it's only been done on-chain. On chain is fairly easy.

**Exchanges** are valuable because they provide and run a matching algorithm and automatic settlement. They also provide a price. The price is usually determined by either the orderbook or the matching algorithm, but most of all price is determined by the state of the system.