# opencx
[![Build Status](https://travis-ci.com/mit-dci/opencx.svg?branch=master)](https://travis-ci.com/mit-dci/opencx)
[![License](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://mit-dci.mit-license.org)

A centralized exchange to help understand what a DEX should be

## Security note
Gosec still detects a bunch of stuff and I need more configuration and documentation for database username and password.

# Requirements
 - Go 1.11+
 - [MariaDB](https://mariadb.org) (not needed for client)

By default the exchange connects to bitcoin, litecoin, and vertcoin testnet nodes.
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

Then start MariaDB:
```sh
sudo systemctl start mariadb
```

Now build and run opencx:
```sh
go build opencx
./opencx
```

# How to run opencx Client
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

---
Image of normal use:
![Picture of normal program use in terminal](../assets/normaluse.png?raw=true)
