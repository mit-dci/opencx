# cxdbsql
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](https://img.shields.io/badge/License-MIT-brightgreen.svg)](https://github.com/mit-dci/opencx/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx/cxdbsql?status.svg)](https://godoc.org/github.com/mit-dci/opencx/cxdbsql)

The cxdbsql packages implements any storage interfaces defined in `cxdb`, using MySQL.
This is currently the most full-featured data store option for OpenCX, implementing all of `OpencxStore`, and most of `OpencxAuctionStore`.
In the future, this should be more easily configurable while still only exporting the abstract `OpencxStore` interface.
