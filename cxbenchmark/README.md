# cxbenchmark

This is a benchmarking suite that I use to benchmark all of the various RPC functions that the server allows.

Specs of computer for benchmarking (Dell XPS 9550 Laptop, obtained from `/proc/cpuinfo` and `dmideinfo`):

OS: Manjaro Linux
```
goos: linux
goarch: amd64
```

```
vendor_id	: GenuineIntel
cpu family	: 6
model		: 94
model name	: Intel(R) Core(TM) i7-6700HQ CPU @ 2.60GHz
cpu MHz		: 799.963
cache size	: 6144 KB
cpu cores	: 4
```

```
Memory Device
	Size: 8192 MB
	Form Factor: SODIMM
	Type: DDR4
	Speed: 2133 MT/s
	Manufacturer: SK Hynix
	Part Number: HMA41GS6AFR8N-TF
	Configured Memory Speed: 2133 MT/s
	Configured Voltage: 1.2 V

Memory Device
	Size: 8192 MB
	Form Factor: SODIMM
	Type: DDR4
	Speed: 2133 MT/s
	Manufacturer: SK Hynix
	Part Number: HMA41GS6AFR8N-TF
	Configured Memory Speed: 2133 MT/s
	Configured Voltage: 1.2 V
```

All tests are run on the regtest environment as well.

# How to run the benchmarks

```sh
go test -v -benchtime TIME -bench=.
```

If you run it in this (`cxbenchmark/`) directory it will create a directory for wallets, in `.benchmarkInfo/`.
By default the benchmarks connect to local Litecoin, Vertcoin, and Bitcoin regtest nodes and there's no configuration for the benchmarks unless you edit the code directly.
If you'd like to do this, the methods `createDefaultParamServerWithKey` and `createDefaultLightServerWithKey` are where the default parameters are set.
It's a little messy because currently both of those methods are in one line each.

These benchmarks currently run one normal exchange server (custodial, proper balances, etc.), and one "light" server.
The "light" server has no settlement, so it's a bit easier to black box benchmark everything else.
What "no settlement" means is that there's no idea of any balance, it just will always accept your order as long as your public key is on a whitelist of "those authorized to make orders," like a traditional stock exchange.

### Currently known limits:

 - From start to finish, with many thousands of blocks, it takes a while to sync up, if you want to use testnet or mainnet for your benchmarks. You probably shouldn't do either of those things, and should use regtest instead for benchmarking.

## Placing orders and matching

At first I tested the performance of order placing, and every time an order was placed it would run matching. I realized this was extremely inefficient, and now run a goroutine to match orders continuously, while taking orders at any point in time. Locks are needed to keep multiple threads from reading and writing to the database, but that is where a distributed database or scaling solution would come in handy. The database could be sharded in the future. A good way to shard would be to separate order rows by price, since the matching for one price is completely separate from another price. This would help scale massively.

Currently the exchange matches orders for all price levels when there are 1000 orders that have been placed since last matching. The number of orders essentially determines how frequently the matching happens. The order book is always updated when an order is placed. There is a method implemented that is able to match for a specific price, but it does not currently run. It could be run after an order is placed.

I am running a benchmark for matching and mass order placement with `go test -v -benchtime 2m -bench=.` in the `cxbenchmark` directory. The test prints out the number of orders were placed after it finishes. Here are the results for multiple runs:

```
dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master ●✚  go test -v -benchtime 2m -bench=.
OCX Client: 2019/02/13 19:39:28 Set up logger
OCX Client: 2019/02/13 19:39:28 Connected to exchange at localhost:12345
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/VariablePlacingAndFilling-8         	     100	1656899711 ns/op
2019/02/13 19:42:15 [INFO] Number of runs: 12120
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	167.366
```
"Transactions per second": 72.416 tx/s

```
 dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master  go test -v -benchtime 2m -bench=.
OCX Client: 2019/02/13 19:42:27 Set up logger
OCX Client: 2019/02/13 19:42:27 Connected to exchange at localhost:12345
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/VariablePlacingAndFilling-8         	     100	1505514559 ns/op
2019/02/13 19:44:58 [INFO] Number of runs: 12120
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	151.776s
```
"Transactions per second": 79.855 tx/s

```
 dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master ●  go test -v -benchtime 2m -bench=.
OCX Client: 2019/02/13 19:45:07 Set up logger
OCX Client: 2019/02/13 19:45:07 Connected to exchange at localhost:12345
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/VariablePlacingAndFilling-8         	     100	1938947412 ns/op
2019/02/13 19:48:23 [INFO] Number of runs: 12120
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	195.528s
```
"Transactions per second": 61.986 tx/s

After testing it more, depending on the size of the orderbook it's either fast or slow. As expected, the more orders for a specific price there are, the more time it takes to match. But once you're done matching many orders it's fast again, since they get deleted. It would be very fast if we didn't care about consistency and only cared that we have proof a valid order was matched.

NOTE: The benchmark requires users `tester` and `othertester` to have a bunch of money.

Here are the results for combination tests. PlaceAndFill test place then immediately fill orders for multiple prices, meaning the matching loop will be somewhat busy. The PlaceBuyThenSell test place many buy orders, then place many sell orders, like an "all at once" operation. I'm still testing whether or not running matching for the price of an order immediately after the order is placed is a good idea, or any slower. The following only run the matching loop. The matching loop can be optimized as well, maybe it should be run on a time increment when the exchange isn't that busy, and run on an order increment when the exchange is busy.

```
 dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master  go test -v -benchtime=10s -bench=.
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/PlaceAndFill1-8         	     200	  72169804 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1-8     	     500	  34813599 ns/op
BenchmarkPlaceOrders/PlaceAndFill10-8        	      30	 671836011 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell10-8    	      50	 344578718 ns/op
BenchmarkPlaceOrders/PlaceAndFill100-8       	       2	5610319672 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell100-8   	       3	4029463074 ns/op
BenchmarkPlaceOrders/PlaceAndFill1000-8      	       1	69659315612 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1000-8  	       1	47541386132 ns/op
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	243.547s
```

```
 dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master  go test -v -benchtime=10s -bench=.
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/PlaceAndFill1-8         	     300	  70524552 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1-8     	     500	  39861595 ns/op
BenchmarkPlaceOrders/PlaceAndFill10-8        	      20	 735612932 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell10-8    	      50	 348971955 ns/op
BenchmarkPlaceOrders/PlaceAndFill100-8       	       2	5652193437 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell100-8   	       3	4086892375 ns/op
BenchmarkPlaceOrders/PlaceAndFill1000-8      	       1	72111150586 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1000-8  	       1	48667777616 ns/op
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	247.990s
```

```
 dan@dan-pc  ~/Documents/Projects/opencx/cxbenchmark   master  go test -v -benchtime=10s -bench=.
goos: linux
goarch: amd64
pkg: github.com/mit-dci/opencx/cxbenchmark
BenchmarkPlaceOrders/PlaceAndFill1-8         	     300	  73141459 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1-8     	     500	  36205554 ns/op
BenchmarkPlaceOrders/PlaceAndFill10-8        	      30	 832876094 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell10-8    	      50	 351879723 ns/op
BenchmarkPlaceOrders/PlaceAndFill100-8       	       2	8249407459 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell100-8   	       3	4381134289 ns/op
BenchmarkPlaceOrders/PlaceAndFill1000-8      	       1	71376749722 ns/op
BenchmarkPlaceOrders/PlaceBuyThenSell1000-8  	       1	47408481982 ns/op
PASS
ok  	github.com/mit-dci/opencx/cxbenchmark	249.438s
```

## Ingesting blocks
Currently when the server starts up, it ingests a whole bunch of blocks, looking for P2PKH outputs to the addresses it controls. When these do not have deposits in them, it is able to process them at about 200 blocks per second.

I currently have a shell script (100deposits.sh) that stress tests the server's ability to take deposits. The full node is actually the bottleneck for testing here.

As for performance goes with the ingests, I've tested 200 100 tx blocks to determine the confirm rate, and here are the results:

```
Test 1
2019/02/12 18:44:00.760648 [INFO] Started ingesting block
2019/02/12 18:44:00.775563 [INFO] Started ingesting block
2019/02/12 18:44:00.794749 [INFO] Done ingesting block
2019/02/12 18:44:00.843491 [INFO] Done ingesting block
2019/02/12 18:44:01.442713 [INFO] Started ingesting block
2019/02/12 18:44:01.566038 [INFO] Done ingesting block
Delta: 0.805390s

Test 2
2019/02/12 18:44:39.517473 [INFO] Started ingesting block
2019/02/12 18:44:39.530320 [INFO] Started ingesting block
2019/02/12 18:44:39.532683 [INFO] Done ingesting block
2019/02/12 18:44:39.565680 [INFO] Started ingesting block
2019/02/12 18:44:39.572690 [INFO] Done ingesting block
2019/02/12 18:44:39.607724 [INFO] Done ingesting block
Delta: 0.090251

Test 3
2019/02/12 18:45:18.222444 [INFO] Started ingesting block
2019/02/12 18:45:18.230069 [INFO] Started ingesting block
2019/02/12 18:45:18.255557 [INFO] Done ingesting block
2019/02/12 18:45:18.273771 [INFO] Done ingesting block
2019/02/12 18:45:18.439655 [INFO] Started ingesting block
2019/02/12 18:45:18.469665 [INFO] Done ingesting block
Delta: 0.247221s

Test 4
2019/02/12 18:45:56.609790 [INFO] Started ingesting block
2019/02/12 18:45:56.615910 [INFO] Started ingesting block
2019/02/12 18:45:56.645177 [INFO] Done ingesting block
2019/02/12 18:45:56.660310 [INFO] Started ingesting block
2019/02/12 18:45:56.663263 [INFO] Done ingesting block
2019/02/12 18:45:56.697782 [INFO] Done ingesting block
Delta: 0.087992s

Test 5
2019/02/12 18:46:35.198364 [INFO] Started ingesting block
2019/02/12 18:46:35.199361 [INFO] Started ingesting block
2019/02/12 18:46:35.232552 [INFO] Done ingesting block
2019/02/12 18:46:35.240114 [INFO] Started ingesting block
2019/02/12 18:46:35.264626 [INFO] Done ingesting block
2019/02/12 18:46:35.297874 [INFO] Done ingesting block
Delta: 0.099510s

Test 6
2019/02/12 18:47:15.003583 [INFO] Started ingesting block
2019/02/12 18:47:15.009319 [INFO] Started ingesting block
2019/02/12 18:47:15.032728 [INFO] Done ingesting block
2019/02/12 18:47:15.051634 [INFO] Done ingesting block
2019/02/12 18:47:15.229780 [INFO] Started ingesting block
2019/02/12 18:47:15.263803 [INFO] Done ingesting block
Delta: 0.260220s

Test 7
2019/02/12 18:47:53.481896 [INFO] Started ingesting block
2019/02/12 18:47:53.493056 [INFO] Started ingesting block
2019/02/12 18:47:53.698393 [INFO] Started ingesting block
2019/02/12 18:47:53.843682 [INFO] Done ingesting block
2019/02/12 18:47:54.817606 [INFO] Done ingesting block
2019/02/12 18:47:54.906257 [INFO] Done ingesting block
Delta: 1.424361s

Test 8
2019/02/12 18:48:32.037255 [INFO] Started ingesting block
2019/02/12 18:48:32.038728 [INFO] Started ingesting block
2019/02/12 18:48:32.166023 [INFO] Done ingesting block
2019/02/12 18:48:32.260861 [INFO] Done ingesting block
2019/02/12 18:48:32.262583 [INFO] Started ingesting block
2019/02/12 18:48:32.383487 [INFO] Done ingesting block
Delta: 0.346232s

Test 9
2019/02/12 18:49:10.467850 [INFO] Started ingesting block
2019/02/12 18:49:10.474062 [INFO] Started ingesting block
2019/02/12 18:49:10.502216 [INFO] Started ingesting block
2019/02/12 18:49:10.578464 [INFO] Done ingesting block
2019/02/12 18:49:10.668538 [INFO] Done ingesting block
2019/02/12 18:49:10.780483 [INFO] Done ingesting block
Delta: 0.312633s

Test 10
2019/02/12 18:49:52.986447 [INFO] Started ingesting block
2019/02/12 18:49:52.993719 [INFO] Started ingesting block
2019/02/12 18:49:53.080049 [INFO] Done ingesting block
2019/02/12 18:49:53.156086 [INFO] Done ingesting block
2019/02/12 18:49:53.238154 [INFO] Started ingesting block
2019/02/12 18:49:53.367143 [INFO] Done ingesting block
Delta: 0.380696s

Test 11
2019/02/12 18:50:43.554055 [INFO] Started ingesting block
2019/02/12 18:50:43.559852 [INFO] Started ingesting block
2019/02/12 18:50:43.709148 [INFO] Done ingesting block
2019/02/12 18:50:43.856247 [INFO] Done ingesting block
2019/02/12 18:50:43.875955 [INFO] Started ingesting block
2019/02/12 18:50:44.014085 [INFO] Done ingesting block
Delta: 0.460030s

Test 12
2019/02/12 18:51:28.940085 [INFO] Started ingesting block
2019/02/12 18:51:28.942794 [INFO] Started ingesting block
2019/02/12 18:51:29.039250 [INFO] Done ingesting block
2019/02/12 18:51:29.120471 [INFO] Done ingesting block
2019/02/12 18:51:29.166521 [INFO] Started ingesting block
2019/02/12 18:51:29.266988 [INFO] Done ingesting block
Delta: 0.326903s

Test 13
2019/02/12 18:52:14.041906 [INFO] Started ingesting block
2019/02/12 18:52:14.046740 [INFO] Started ingesting block
2019/02/12 18:52:14.146856 [INFO] Done ingesting block
2019/02/12 18:52:14.219225 [INFO] Done ingesting block
2019/02/12 18:52:14.269570 [INFO] Started ingesting block
2019/02/12 18:52:14.365146 [INFO] Done ingesting block
Delta: 0.323240s

Test 14
2019/02/12 18:52:52.236238 [INFO] Started ingesting block
2019/02/12 18:52:52.239431 [INFO] Started ingesting block
2019/02/12 18:52:52.275271 [INFO] Started ingesting block
2019/02/12 18:52:52.370295 [INFO] Done ingesting block
2019/02/12 18:52:52.449486 [INFO] Done ingesting block
2019/02/12 18:52:52.551128 [INFO] Done ingesting block
Delta: 0.314890s

Test 15
2019/02/12 18:53:30.713974 [INFO] Started ingesting block
2019/02/12 18:53:30.727448 [INFO] Started ingesting block
2019/02/12 18:53:30.810519 [INFO] Done ingesting block
2019/02/12 18:53:30.908802 [INFO] Done ingesting block
2019/02/12 18:53:30.965204 [INFO] Started ingesting block
2019/02/12 18:53:31.056883 [INFO] Done ingesting block
Delta: 0.342909s

Test 16
2019/02/12 18:54:12.252197 [INFO] Started ingesting block
2019/02/12 18:54:12.257574 [INFO] Started ingesting block
2019/02/12 18:54:12.349950 [INFO] Done ingesting block
2019/02/12 18:54:12.430650 [INFO] Done ingesting block
2019/02/12 18:54:12.474825 [INFO] Started ingesting block
2019/02/12 18:54:12.573271 [INFO] Done ingesting block
Delta: 0.321074s

Test 17
2019/02/12 18:54:53.256631 [INFO] Started ingesting block
2019/02/12 18:54:53.258374 [INFO] Started ingesting block
2019/02/12 18:54:53.292427 [INFO] Started ingesting block
2019/02/12 18:54:53.345431 [INFO] Done ingesting block
2019/02/12 18:54:53.505007 [INFO] Done ingesting block
2019/02/12 18:54:53.639642 [INFO] Done ingesting block
Delta: 0.383011s

Test 18
2019/02/12 18:55:41.545643 [INFO] Started ingesting block
2019/02/12 18:55:41.547652 [INFO] Started ingesting block
2019/02/12 18:55:41.805696 [INFO] Done ingesting block
2019/02/12 18:55:41.881197 [INFO] Done ingesting block
2019/02/12 18:55:41.907357 [INFO] Started ingesting block
2019/02/12 18:55:42.030403 [INFO] Done ingesting block
Delta: 0.484760s

Test 19
2019/02/12 18:56:24.851157 [INFO] Started ingesting block
2019/02/12 18:56:24.852497 [INFO] Started ingesting block
2019/02/12 18:56:24.955692 [INFO] Done ingesting block
2019/02/12 18:56:25.045376 [INFO] Done ingesting block
2019/02/12 18:56:25.079247 [INFO] Started ingesting block
2019/02/12 18:56:25.209738 [INFO] Done ingesting block
Delta: 0.358581s

Test 20
2019/02/12 18:57:06.592153 [INFO] Started ingesting block
2019/02/12 18:57:06.647205 [INFO] Started ingesting block
2019/02/12 18:57:06.690890 [INFO] Done ingesting block
2019/02/12 18:57:06.763172 [INFO] Done ingesting block
2019/02/12 18:57:06.900448 [INFO] Started ingesting block
2019/02/12 18:57:07.015242 [INFO] Done ingesting block
Delta: 0.423089s

Sum of deltas: 7.466090s
Average time per test: 0.3733045s
Transactions per test: 300
"Transactions per second": 803
```
