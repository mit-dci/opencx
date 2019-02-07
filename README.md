# opencx
A centralized exchange to help understand what a DEX should be

# Requirements
 - Go 1.11+
 - [MariaDB](https://mariadb.org) (not needed for client)

## Optional - If you'd like to play with and mine your own money to trade
 - A running bitcoin regtest node on the local machine
 - A running litecoin regtest node on the local machine
 - A running vertcoin regtest node on the local machine

It's all the same protocol, so it should work with btcd and whatnot (not vtcd though, the address prefix on that is wrong), but I've been testing with [bitcoind](https://github.com/bitcoin/bitcoin), [litecoind](https://github.com/litecoin-project/litecoin), and [vertcoind](https://github.com/vertcoin-project/vertcoin-core). 

I'm still working on adding an option to indicate that you're running a regtest node, so the default right now is just connect to testnets.

Here is how I have my configs set up, you should copy and paste these for easy set up:

##### `~/.bitcoin/bitcoin.conf`
```
server=1
regtest=1
daemon=1
deprecatedrpc=generate
debug=net
```

##### `~/.vertcoin/vertcoin.conf`
```
server=1
daemon=1
regtest=1
# Because vertcoin sets default to the same as bitcoin :(
port=20444
rpcport=20443
debug=net
```

##### `~/.litecoin/litecoin.conf`
```
server=1
daemon=1
regtest=1
debug=net
```

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

Image of normal use:
![Picture of normal program use in terminal](../assets/normaluse.png?raw=true)
