# Decentralized exchange tech survey

I'm going to try to review as much as possible about decentralized exchange technology

## Research, proposed DEXes, anything not currently implemented
It's good to get a survey of the research and not-yet-implemented solutions to problems that people supposedly have with exchange. Some things here are DEXes, some are techniques used to maybe decrease trust needed in exchange, but most are not formal and not peer reviewed.
 - [ ] Arwen
 - [ ] Nash / NEX
   - [ ] [NEX Whitepaper](assets/whitepaper_v2.pdf)
   - [ ] [The heart of Nash: Our off-chain matching engine](https://medium.com/nashsocial/the-heart-of-nash-our-off-chain-matching-engine-499cf2c23670)
 - [ ] [gnosis/dex-research](https://github.com/gnosis/dex-research)
   - [ ] [dFusion](https://github.com/gnosis/dex-research/tree/master/dFusion)
   - [ ] [Batch Auction Optimization](https://github.com/gnosis/dex-research/tree/master/BatchAuctionOptimization)
   - [ ] [Plasma Research](https://github.com/gnosis/dex-research/tree/master/PlasmaResearch)
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
 - [ ] Binance Chain
 - [ ] Altcoin.io - [Another Plasma DEX](https://blog.altcoin.io/plasma-dex-v1-launching-next-month-4cb5e5ea56f6)

## Topics
These topics are for me to review and asses what the benefits of each are and how they would be relevant in an exchange. Most of these are reviewed topics as well as used for more general things.
 - [ ] "Provably fair" matching
   - [ ] Difference between verifiable computation and zero knowledge proofs
 - [ ] Proof of assets
 - [ ] Cross chain swaps
 - [ ] Exchange channels (different than cross chain - like arwen?)

## Other articles about crypto exchanges in general, mostly greivances and events
Looking at history is important, especially because we may have seen some of the problems that people care about.
 - [ ] [Bitcoin exchanges may not be ready for the big time -- Quartz](https://qz.com/1120991/bitcoin-exchanges-may-not-be-ready-for-the-big-time/)
 - [ ] [The History of the Mt Gox Hack: Bitcoin's Biggest Heist](https://blockonomi.com/mt-gox-hack/)
 - [ ] [A crypto exchange CEO dies - With the only key to $137 million -- WIRED](https://www.wired.com/story/crypto-exchange-ceo-dies-holding-only-key/)
 - [ ] [The State of Decentralized Exchanges](https://hackernoon.com/the-state-of-decentralized-exchanges-235064446ab0)

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

## Centralized exchanges
It's also good to see what centralized exchanges could do if they were malicious - exploring attack vectors on the user and exchange.
 - [ ] OpenCX (lol)
 - [ ] Coinbase
 - [ ] Binance
 - [ ] MtGox
 - [ ] Bitfinex
 - [ ] Gemini
 - [ ] Smaller exchanges

## Implemented DEXes and exchange problem solutions
These are currently implemented "solutions" to problems that users supposedly have with cryptocurrency exchanges. I'll determine whether or not these actually solve any problems, and how well they solve them. One thing that I will be covering a lot is whether or not they are platforms which are bound to a single cryptocurrency, and rely on the fact that said currency is the biggest in order to solve what they set out to solve.
 - [ ] BitShares
 - [ ] Binance Chain (after Feb 20th)
 - [ ] 0x
 - [ ] Kyber Network
 - [ ] EtherDelta
 - [ ] IDEX
 - [ ] Uniswap

# Research, proposed DEXes, anything not currently implemented

## Arwen

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

### dFusion
dFusion is sort of a smart contract that will take in a bunch of orders, freeze them, then have people propose which orders to match using zk-SNARKS. Not implemented yet. [Here's a link to the ethresear.ch post](https://ethresear.ch/t/building-a-decentralized-exchange-using-snarks/3928), where they say that each snark will cost $1200 to produce on AWS. The incentive to spend $1200 would be that you get trading fees from the trades. They try to implement their batch auction optimization too, which makes since, it's literally one giant batch trade. They say auction closing / matching times would be anywhere from 3 hours to 1 day. The matching optimizes for "trader surplus" or "trading volume". The high cost of calculating the SNARK is a "crypto-economic" incentive to only post valid solutions. They also collect orders on chain, although they would probably do it on a plasma chain, so it's still ethereum only.

### Batch Auction Optimization
This is basically a way of making sure low liquidity tokens are sorta liquid.


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

## Altcoin.io

# Topics

## "Provably fair" matching

## Proof of assets

## Cross chain swaps

## Exchange channels (different than cross chain - like arwen?)

# Other articles about crypto exchanges in general, mostly greivances and events

## Bitcoin exchanges may not be ready for the big time -- Quartz

## The History of the Mt Gox Hack: Bitcoin's Biggest Heist

## A crypto exchange CEO dies - With the only key to $137 million -- WIRED

## The State of Decentralized Exchanges

# Centralized Exchanges

## OpenCX (lol)

## Coinbase

## Binance

## MtGox

## Bitfinex

## Gemini

## Smaller exchanges

# Implemented DEXes and exchange problem solutions

## BitShares

## Binance Chain (after Feb 20th)

## 0x

## Kyber Network

## EtherDelta

## IDEX

## Uniswap