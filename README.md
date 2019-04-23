# opencx
[![Build Status](https://travis-ci.org/mit-dci/opencx.svg?branch=master)](https://travis-ci.org/mit-dci/opencx)
[![License](http://img.shields.io/:license-MIT-000.svg)](http://mit-dci.mit-license.org)
[![GoReport](https://goreportcard.com/badge/github.com/mit-dci/opencx)](https://goreportcard.com/badge/github.com/mit-dci/opencx)
[![GoDoc](https://godoc.org/github.com/mit-dci/opencx?status.svg)](https://godoc.org/github.com/mit-dci/opencx)

OpenCX stands for Open Centralized eXchange, it's an open-source cryptocurrency exchange toolkit originally built to help understand what a decentralized exchange should be. It's meant to be modular enough so features that increase trustlessness in cryptocurrency exchange can be implemented and experimented with. Included are packages for lightning support, RPC, authentication via the NOISE protocol, and a Golang API supporting multiple forms of authentication.

# Requirements
 - Go 1.11+
 - A MySQL Database (not needed for client)

# Demo

![gif of program in normal use](../assets/opencxdemo.gif?raw=true)

# How to run opencx server / exchange
First clone the repo and install dependencies:
```sh
git clone git@github.com/mit-dci/opencx.git
cd opencx
go get
```

Then start MariaDB, or any other MYSQL database:
```sh
sudo systemctl start mariadb
```

Now build and run opencx:
```sh
go build opencx
./opencx
```

# How to run the opencx command line client
Clone the repo and install dependencies as in the steps above:
```sh
git clone git@github.com/mit-dci/opencx
go get
```

Now build the binary:
```sh
cd cmd/ocx
go build
```

You can now issue any of the commands in the cxrpc README.md file.

# Configuration
There are configuration options (both command line and .conf) for the client and the server, and by default home folders for these files will be created at `~/.opencx` and `~/.ocx` respectively. You can decide whether or not to use the NOISE protocol for authentication, which hostnames and ports to use for connecting to certain clients, which coins you would like to support, and whether or not to support lightning.

If you'd like to add your own coins, just add a coinparam struct like in `lit`.
