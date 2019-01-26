# opencx
A centralized exchange to help understand what a DEX should be

# Requirements
 - Go 1.11+
 - [MariaDB](https://mariadb.org) (not needed for client)

# How to run opencx server / exchange
First clone the repo and install dependencies:
```sh
git clone git@github.com/mit-dci/opencx
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
