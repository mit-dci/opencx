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
 - [ ] ethresear.ch ["Decentralized exchanges" category](https://ethresear.ch/c/decentralized-exchanges) as of 2/19/2019
   - [x] [Introducing DTRADE. Confoederation dapp](https://ethresear.ch/t/introducing-dtrade-confoederatio-dapp/4870)
   - [ ] [Self-referential proxy exchange - does this exist?](https://ethresear.ch/t/self-referential-proxy-exchange-does-this-exist/4515)
   - [ ] [Batch auctions with uniform clearing price on plasma](https://ethresear.ch/t/batch-auctions-with-uniform-clearing-price-on-plasma/2554)
   - [ ] [DutchX - fully decentralized auction based exchange](https://ethresear.ch/t/dutchx-fully-decentralized-auction-based-exchange/2443)
   - [ ] [A note for the Dutch Team & other Dapp developers](https://ethresear.ch/t/a-note-for-the-dutch-team-other-dapp-developers/2637)
   - [ ] [Improving front-running resistance of x*y=k market makers](https://ethresear.ch/t/improving-front-running-resistance-of-x-y-k-market-makers/1281)
   - [ ] [Limit orders and slippage resistance in x*y=k market makers](https://ethresear.ch/t/limit-orders-and-slippage-resistance-in-x-y-k-market-makers/2071)
   - [ ] [Reverse Parimutuel Options on Bitcoin](https://ethresear.ch/t/reverse-parimutuel-options-on-bitcoin/1816)
 - [ ] Binance Chain
 - [ ] Altcoin.io - [Another Plasma DEX](https://blog.altcoin.io/plasma-dex-v1-launching-next-month-4cb5e5ea56f6)

## Topics
These topics are for me to review and asses what the benefits of each are and how they would be relevant in an exchange. Most of these are reviewed topics as well as used for more general things.
 - [ ] "Provably fair" matching
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

This is one of the most thoughtful pieces of decentralized exchange research, since they fully recognize that the matching engine will have to be trusted, and count it as a trade-off. The efforts towards making the matching less centralized are also thoughtful.

### The Heart of Nash Article review and thoughts

## Gnosis DEX Research
This is a repo by gnosis called dex-research so I decided to check it out.

### dFusion
dFusion is sort of a smart contract that will take in a bunch of orders, freeze them, then have people propose which orders to match using zk-SNARKS. Not implemented yet. [Here's a link to the ethresear.ch post](https://ethresear.ch/t/building-a-decentralized-exchange-using-snarks/3928), where they say that each snark will cost $1200 to produce on AWS. The incentive to spend $1200 would be that you get trading fees from the trades. They try to implement their batch auction optimization too, which makes since, it's literally one giant batch trade. They say auction closing / matching times would be anywhere from 3 hours to 1 day. The matching optimizes for "trader surplus" or "trading volume". The high cost of calculating the SNARK is a "crypto-economic" incentive to only post valid solutions. They also collect orders on chain, although they would probably do it on a plasma chain, so it's still ethereum only.

### Batch Auction Optimization
This is basically a way of making sure low liquidity tokens are sorta liquid.

### Plasma Research
This part is just TeX but it defines how you would do batch auction stuff on plasma. Again, not really something that would be interesting for someone making a decentralized exchange that isn't just an erc20 token swap contract on ethereum.

## ethresear.ch DEX topics
The ethresear.ch DEX category seems to be pretty weak when it comes to talking about actual decentralized exchange, mostly just proposals for how to scale a DEX that is only on ethereum, or posts about a "new DEX Dapp". Now for the reviews:

### DTRADE / Confoederation Dapp
This post is really just a link to a blog post about a platform called "Confoederation TRADE," which, from reading the actual post as well as the website, seems to be a collection of smart contracts on ethereum that are neither implemented nor justified. Reading the blog post and the website was a waste of time.

### Self-referential proxy exchange - does this exist?

### Batch auctions with uniform clearing price on plasma

### DutchX - fully decentralized auction based exchange

### A note for the Dutch Team & other Dapp developers

### Improving front-running resistance of x*y=k market makers

### Limit orders and slippage resistance in x*y=k market makers

### Reverse Parimutuel Options on Bitcoin

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