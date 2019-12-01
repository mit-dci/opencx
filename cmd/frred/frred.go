package main

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rjected/lit/coinparam"
	"github.com/Rjected/lit/crypto/koblitz"

	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/opencx/cxauctionrpc"
	"github.com/mit-dci/opencx/cxauctionserver"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

type frredConfig struct {
	ConfigFile string

	// stuff for files and directories
	LogFilename  string `long:"logFilename" description:"Filename for output log file"`
	FrredHomeDir string `long:"dir" description:"Location of the root directory relative to home directory"`

	// stuff for ports
	Rpcport uint16 `short:"p" long:"rpcport" description:"Set RPC port to connect to"`
	Rpchost string `long:"rpchost" description:"Set RPC host to listen to"`

	// logging and debug parameters
	LogLevel []bool `short:"v" description:"Set verbosity level to verbose (-v), very verbose (-vv) or very very verbose (-vvv)"`

	// logging for lit nodes (find something better than w)
	LitLogLevel []bool `short:"w" description:"Set verbosity level to verbose (-w), very verbose (-ww) or very very verbose (-www)"`

	// Resync?
	Resync bool `short:"r" long:"resync" description:"Do you want to resync all chains?"`

	// networks that we can connect to
	Vtchost     string `long:"vtc" description:"Connect to Vertcoin full node. Specify a socket address."`
	Btchost     string `long:"btc" description:"Connect to bitcoin full node. Specify a socket address."`
	Ltchost     string `long:"ltc" description:"Connect to a litecoin full node. Specify a socket address."`
	Tn3host     string `long:"tn3" description:"Connect to bitcoin testnet3. Specify a socket address."`
	Lt4host     string `long:"lt4" description:"Connect to litecoin testnet4. Specify a socket address."`
	Tvtchost    string `long:"tvtc" description:"Connect to Vertcoin test node. Specify a socket address."`
	Reghost     string `long:"reg" description:"Connect to bitcoin regtest. Specify a socket address."`
	Litereghost string `long:"litereg" description:"Connect to litecoin regtest. Specify a socket address."`
	Rtvtchost   string `long:"rtvtc" description:"Connect to Vertcoin regtest node. Specify a socket address."`

	// configuration for concurrent RPC users.
	MaxPeers    uint16 `long:"numpeers" description:"Maximum number of peers that you'd like to support"`
	MinPeerPort uint16 `long:"minpeerport" description:"Port to start creating ports for peers at"`
	Lithost     string `long:"lithost" description:"Host for the lightning node on the exchange to run"`
	Litport     uint16 `long:"litport" description:"Port for the lightning node on the exchange to run"`

	// filename for key
	KeyFileName string `long:"keyfilename" short:"k" description:"Filename for private key within root opencx directory used to send transactions"`

	// auth or unauth rpc?
	AuthenticatedRPC bool `long:"authrpc" description:"Whether or not to use authenticated RPC"`

	// support lightning or not to support lightning?
	LightningSupport bool `long:"lightning" description:"Whether or not to support lightning on the exchange"`

	// Auction server options
	AuctionTime  uint64 `long:"auctiontime" description:"Time it should take to generate a timelock puzzle protected order"`
	MaxBatchSize uint64 `long:"maxbatchsize" description:"Maximum number of orders that can go in a batch"`
}

var (
	defaultHomeDir = os.Getenv("HOME")

	// used as defaults before putting into parser
	defaultfrredHomeDirName = defaultHomeDir + "/.opencx/frred/"
	defaultRpcport          = uint16(12345)
	defaultRpchost          = "localhost"
	defaultMaxPeers         = uint16(64)
	defaultMinPeerPort      = uint16(25565)
	defaultLithost          = "localhost"
	defaultLitport          = uint16(12346)

	// Yes we want to use noise-rpc
	defaultAuthenticatedRPC = true

	// Yes we want lightning
	defaultLightningSupport = true

	// default auction options
	defaultAuctionTime  = uint64(30000)
	defaultMaxBatchSize = uint64(1000)
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *frredConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

func main() {
	var err error

	conf := frredConfig{
		FrredHomeDir:     defaultfrredHomeDirName,
		Rpcport:          defaultRpcport,
		Rpchost:          defaultRpchost,
		MaxPeers:         defaultMaxPeers,
		MinPeerPort:      defaultMinPeerPort,
		Lithost:          defaultLithost,
		Litport:          defaultLitport,
		AuthenticatedRPC: defaultAuthenticatedRPC,
		LightningSupport: defaultLightningSupport,
		AuctionTime:      defaultAuctionTime,
		MaxBatchSize:     defaultMaxBatchSize,
	}

	// Check and load config params
	key := opencxSetup(&conf)

	// Generate the coin list based on the parameters we know
	coinList := generateCoinList(&conf)

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		logging.Fatalf("Could not generate asset pairs from coin list: %s", err)
	}

	// Create in memory matching engine
	var mengines map[match.Pair]match.AuctionEngine
	if mengines, err = cxdbsql.CreateAuctionEngineMap(pairList); err != nil {
		logging.Fatalf("Error creating auction engines for pairs: %s", err)
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbsql.CreateSettlementEngineMap(coinList); err != nil {
		logging.Fatalf("Error creating settlement engine map: %s", err)
	}

	var auctionBooks map[match.Pair]match.AuctionOrderbook
	if auctionBooks, err = cxdbsql.CreateAuctionOrderbookMap(pairList); err != nil {
		logging.Fatalf("Error creating auction orderbook map: %s", err)
	}

	var puzzleStores map[match.Pair]cxdb.PuzzleStore
	if puzzleStores, err = cxdbsql.CreatePuzzleStoreMap(pairList); err != nil {
		logging.Fatalf("Error creating puzzle store map: %s", err)
	}

	var batchers map[match.Pair]match.AuctionBatcher
	if batchers, err = cxauctionserver.CreateAuctionBatcherMap(pairList, conf.MaxBatchSize); err != nil {
		logging.Fatalf("Error creating batcher map: %s", err)
	}

	// Anyways, here's where we set the server
	var frredServer *cxauctionserver.OpencxAuctionServer
	if frredServer, err = cxauctionserver.InitServer(setEngines, mengines, auctionBooks, puzzleStores, batchers, 100, conf.AuctionTime); err != nil {
		logging.Fatalf("Error initializing server: \n%s", err)
	}

	if err = frredServer.StartClockRandomAuction(); err != nil {
		logging.Fatalf("Error starting clock: %s", err)
	}

	// Register RPC Commands and set server
	var rpcListener *cxauctionrpc.AuctionRPCCaller
	if rpcListener, err = cxauctionrpc.CreateRPCForServer(frredServer); err != nil {
		logging.Fatalf("Error creating rpc caller for server: %s", err)
	}

	// SIGINT and SIGTERM and SIGQUIT handler for CTRL-c, KILL, CTRL-/, etc.
	go func() {
		logging.Infof("Notifying signals")
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGQUIT)
		signal.Notify(sigs, syscall.SIGTERM)
		signal.Notify(sigs, syscall.SIGINT)
		for {
			signal := <-sigs
			logging.Infof("Received %s signal, Stopping server gracefully...", signal.String())

			// send off button to off button
			if err = rpcListener.KillServerNoWait(); err != nil {
				logging.Fatalf("Error killing server: %s", err)
			}

			return
		}
	}()

	if !conf.AuthenticatedRPC {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error listening for rpc for auction serer: %s", err)
		}
	} else {
		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error listening for noise rpc for auction serer: %s", err)
		}
	}

	// wait until the listener dies
	rpcListener.WaitUntilDead()

	return
}
