# opencx
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](http://img.shields.io/badge/License-MIT-brightgreen.svg)](./LICENSE)
> [![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx?status.svg)](https://godoc.org/github.com/mit-dci/opencx)

Cryptocurrency exchanges are some of the largest businesses in the cryptocurrency space, and their reserves are often viewed as "honeypots" of cryptocurrencies.
Because of this, cryptocurrency exchanges have been a hotbed of crime in the form of hacks, front-running, wash trading, fake orderbooks, and much more.
In order for cryptocurrency to be successful, we need safe, trustworthy ways to exchange cryptocurrencies, without fear that coins will be stolen, or trades executed unfairly.
Additionally, the vast majority of exchange software is closed-source, and exchanges have historically not implemented technological upgrades that would substantially decrease risk for users.

OpenCX hopes to solve this problem by making it trivially easy to run a secure, scalable cryptocurrency exchange which implements many of these features, including:

  - Layer two compatibility
  - Non-custodial exchange
  - Anti front-running
  - Public orderbook auditability
  - *More to come...*

Additionally, all of the components of OpenCX are designed to be swappable, secure, and scalable.
The goal is to fit those requirements and implement features similar to that of modern cryptocurrency exchanges, while keeping high quality software.

**DO NOT use in a production environment, this project is a work in progress!**

*Pull requests and issues encouraged!*

# Contributing

Please see the 
[contributing](./CONTRIBUTING.md) 
file to get started with contributing!

# Requirements
 - Go 1.12+
 - A MySQL Database (not needed for client)

# Demo

![gif of program in normal use](../assets/opencxdemo.gif?raw=true)

# How to run opencx server / exchange
First clone the repo and install dependencies:
```sh
git clone git@github.com/mit-dci/opencx.git
cd opencx
go get -v ./...
```

Then start MariaDB, or any other MYSQL database:
```sh
sudo systemctl start mariadb
```

Now build and run opencx:
```sh
go build ./cmd/opencxd/...
./opencxd
```

# How to run the opencx command line client
Clone the repo and install dependencies as in the steps above:
```sh
git clone git@github.com/mit-dci/opencx
cd opencx
go get -v ./...
```

Now build the binary:
```sh
go build ./cmd/ocx/...
./ocx
```

You can now issue any of the commands in the cxrpc README.md file.

# Configuration
There are configuration options (both command line and .conf) for the client and the server, and by default home folders for these files will be created at `~/.opencx/opencxd/` and `~/.opencx/ocx/` respectively. You can decide whether or not to use the 
[NOISE protocol](http://www.noiseprotocol.org/)
for authentication, which hostnames and ports to use for connecting to certain clients, which coins you would like to support, and whether or not to support lightning.

If you'd like to add your own coins, just add a coinparam struct like in `lit`.
