# opencx
[![Build Status](https://img.shields.io/travis/com/mit-dci/opencx.svg?style=flat-square)](https://travis-ci.com/mit-dci/opencx)
[![License](http://img.shields.io/:license-MIT-000.svg?style=flat-square)](http://mit-dci.mit-license.org)

A centralized exchange to help understand what a DEX should be.

## Documentation
[You can see the godoc documentation here!](https://godoc.org/github.com/mit-dci/opencx)

## Security note
Gosec still detects a bunch of stuff.

# Requirements
 - Go 1.11+
 - [MariaDB](https://mariadb.org) (not needed for client)

There are configuration files that will let you run more than just bitcoin, litecoin, and vertcoin testnet.
You can now start an exchange that is running bitcoin, litecoin, and vertcoin mainnet alongside regtest and testnet.
I'm working on adding in lightning compatibility for all of this.
If you'd like to add your own coins, just add them to the coinparams in `lit`.

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
There are configuration options (both command line and .conf) for the client and the server, and by default home folders for these files will be created at `~/.opencx` and `~/.ocx` respectively. You can decide whether or not to use the NOISE protocol for authentication, which hostnames and ports to use for connecting to certain clients

---
Image of normal use:
![Picture of normal program use in terminal](../assets/normaluse.png?raw=true)
