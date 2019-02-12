# cxbenchmark

This is a benchmarking suite that I use to benchmark all of the various RPC functions that the server allows.

Specs of computer for benchmarking (Dell XPS 9550 Laptop, obtained from `/proc/cpuinfo` and `dmideinfo`):
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

## Currently known limits:

 - If you try to use SQL injection you will succeed. The honor system is currently in place to protect against that vulnerability.

#### Ingesting blocks
Currently when the server starts up, it ingests a whole bunch of blocks, looking for P2PKH outputs to the addresses it controls. When these do not have deposits in them, it is able to process them at about 200 blocks per second.

I currently have a shell script (100deposits.sh) that stress tests the server's ability to take deposits. The full node is actually the bottleneck for testing here.