# opencx

[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](http://img.shields.io/badge/License-MIT-brightgreen.svg)](./LICENSE)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx?status.svg)](https://godoc.org/github.com/mit-dci/opencx)
<!-- [![Go Report Card](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/report/github.com/mit-dci/opencx) -->

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

## Demo

![gif of program in normal use](../assets/opencxdemo.gif?raw=true)

## Contributing

Please see the
[contributing](./CONTRIBUTING.md)
file to get started with contributing!

# Setup

## Requirements

- Go 1.12+
- A MySQL Database (not needed for client)
- GMP (GNU Multiple Precision Arithmetic Library)

## Installing

### Installing GMP

#### Debian

```sh
sudo apt-get install libgmp3-dev
```

#### macOS

```sh
brew install gmp
```

### Clone repo and install dependencies

```sh
git clone git@github.com/mit-dci/opencx.git
cd opencx
go get -v ./...
```

## Running opencx server / exchange

You will need to run MariaDB or any other MySQL database in-order to run the server. You can configure authentication details for your database at `~/.opencx/db/sqldb.conf`

### Start your database (MariaDB in this case)

#### Linux

```sh
sudo systemctl start mariadb
```

#### macOS

```sh
mysql.server start
```

### Now build and run opencx

```sh
go build ./cmd/opencxd/...
./opencxd
```

## Running opencx CLI client

```sh
go build ./cmd/ocx/...
./ocx
```

You can now issue any of the commands in the cxrpc README.md file.

## Configuration

There are configuration options (both command line and .conf) for the client and the server, and by default home folders for these files will be created at `~/.opencx/opencxd/` and `~/.opencx/ocx/` respectively. You can decide whether or not to use the
[NOISE protocol](http://www.noiseprotocol.org/)
for authentication, which hostnames and ports to use for connecting to certain clients, which coins you would like to support, and whether or not to support lightning.

If you'd like to add your own coins, just add a coinparam struct like in `lit`.
