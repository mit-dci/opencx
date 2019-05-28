# cxdb
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://github.com/mit-dci/opencx/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx/cxdb?status.svg)](https://godoc.org/github.com/mit-dci/opencx/cxdb)

The package cxdb contains two interfaces in order to standardize and abstract data store interaction.
There are also contained implementations of these interfaces, some more complete than others.
These interfaces are `OpencxStore` and `OpencxAuctionStore`, the former defining interfaces for a typical exchange, the other based on batch auctions.

Here are the respective statuses of implementations:

  - **cxdbmemory**
    - Conforms to the `OpencxAuctionStore` interface but some of the methods are not functional.
    - This is, and always will be, **ONLY** for rapid prototyping, **NOT** for actual use.
  - **cxdbredis**
    - Defunct, not updated, doesn't work.
    - The reason why it ever existed is because that is what OpenCX used before the need for ACID transactions was realized.
    - See [Issue #16](https://github.com/mit-dci/opencx/issues/16).
  - **cxdbsql**
    - Conforms to `OpencxStore` and `OpencxAuctionStore`.
    - There is a lot of code here, and it needs a lot of tests, as well as a refactor.

The issues related to refactoring cxdb are [#16](https://github.com/mit-dci/opencx/issues/16), [#11](https://github.com/mit-dci/opencx/issues/11), [#6](https://github.com/mit-dci/opencx/issues/6), [#7](https://github.com/mit-dci/opencx/issues/7).
