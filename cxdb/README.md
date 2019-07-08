# cxdb
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://github.com/mit-dci/opencx/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx/cxdb?status.svg)](https://godoc.org/github.com/mit-dci/opencx/cxdb)

The code which manages exchange datastore interactions is split into a set of useful interfaces:
### SettlementEngine
SettlementEngine has two methods. One method checks whether or not a settlement execution could take place if it were executed (think of this as "can we credit this user X of an asset").
The second method actually executes the settlement execution(s).

### AuctionEngine
AuctionEngine is the matching engine for auction orders. It has a place method, a cancel method, and a match method. The match method takes an auction ID as input, since orders cannot be matched cross-auction. This matches according to a clearing price based auction matching algorithm.
### LimitEngine
LimitEngine is the matching engine for limit orders. It has a place method, a cancel method, and a match method. This matches according to a price-time priority auction matching algorithm.
### AuctionOrderbook
AuctionOrderbook gets updated by the auction matching engine, and is viewable by the user. This also has other methods that may be useful, for example methods to get orders by ID, pubkey, or auction.
### LimitOrderbook
LimitOrderbook is very similar to AuctionOrderbook except it does not have methods dependent on a specific auction, since limit orderbooks do not have auctions.
### PuzzleStore
PuzzleStore is a simple store for storing timelock puzzles, as well as marking specific timelock puzzles to commit to or match.
### DepositStore
DepositStore stores the mapping from pubkey to deposit address. This also keeps track of pending deposits. Pending deposits do not have a fixed number of confirmations, and can be set arbitrarily.

### DB interface implementation status
  - SettlementEngine
    - [x] cxdbsql
    - [x] cxdbmemory
    - [ ] cxdbredis
  - AuctionEngine
    - [x] cxdbsql
    - [ ] cxdbemory (partial)
    - [ ] cxdbredis
  - LimitEngine
    - [x] cxdbsql
    - [ ] cxdbmemory
    - [ ] cxdbredis
  - AuctionOrderbook
    - [x] cxdbsql
    - [ ] cxdbmemory
    - [ ] cxdbredis
  - LimitOrderbook
    - [x] cxdbsql
    - [ ] cxdbmemory
    - [ ] cxdbredis
  - PuzzleStore
    - [x] cxdbsql
    - [x] cxdbmemory
    - [ ] cxdbredis
  - DepositStore
    - [x] cxdbsql
    - [ ] cxdbmemory
    - [ ] cxdbredis

Some old code still exists in `cxdbmemory`.
The issues related to refactoring cxdb are [#16](https://github.com/mit-dci/opencx/issues/16).
