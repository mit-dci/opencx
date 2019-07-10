package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"

	flags "github.com/jessevdk/go-flags"
	util "github.com/mit-dci/opencx/chainutils"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbsql"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/cxserver"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

type opencxConfig struct {
	ConfigFile string

	// stuff for files and directories
	LogFilename   string `long:"logFilename" description:"Filename for output log file"`
	OpencxHomeDir string `long:"dir" description:"Location of the root directory for opencxd"`

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
}

var (
	defaultHomeDir = os.Getenv("HOME")

	// used as defaults before putting into parser
	defaultOpencxHomeDirName = defaultHomeDir + "/.opencx/opencxd/"
	defaultRpcport           = uint16(12345)
	defaultRpchost           = "localhost"
	defaultMaxPeers          = uint16(64)
	defaultMinPeerPort       = uint16(25565)
	defaultLithost           = "localhost"
	defaultLitport           = uint16(12346)

	// Yes we want to use noise-rpc
	defaultAuthenticatedRPC = true

	// Yes we want lightning
	defaultLightningSupport = true
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *opencxConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

func main() {
	var err error

	conf := opencxConfig{
		OpencxHomeDir:    defaultOpencxHomeDirName,
		Rpcport:          defaultRpcport,
		Rpchost:          defaultRpchost,
		MaxPeers:         defaultMaxPeers,
		MinPeerPort:      defaultMinPeerPort,
		Lithost:          defaultLithost,
		Litport:          defaultLitport,
		AuthenticatedRPC: defaultAuthenticatedRPC,
		LightningSupport: defaultLightningSupport,
	}

	// Check and load config params
	key := opencxSetup(&conf)

	// Generate the coin list based on the parameters we know
	coinList := generateCoinList(&conf)

	var pairList []*match.Pair
	if pairList, err = match.GenerateAssetPairs(coinList); err != nil {
		logging.Fatalf("Could not generate asset pairs from coin list: %s", err)
	}

	var mengines map[match.Pair]match.LimitEngine
	if mengines, err = cxdbsql.CreateLimitEngineMap(pairList); err != nil {
		logging.Fatalf("Error creating limit engine map with coinlist for opencxd: %s", err)
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if setEngines, err = cxdbsql.CreateSettlementEngineMap(coinList); err != nil {
		logging.Fatalf("Error creating settlement engine map for opencxd: %s", err)
	}

	var limBooks map[match.Pair]match.LimitOrderbook
	if limBooks, err = cxdbsql.CreateLimitOrderbookMap(pairList); err != nil {
		logging.Fatalf("Error creating limit orderbook map for opencxd: %s", err)
	}

	var depositStores map[*coinparam.Params]cxdb.DepositStore
	if depositStores, err = cxdbsql.CreateDepositStoreMap(coinList); err != nil {
		logging.Fatalf("Error creating deposit store map for opencxd: %s", err)
	}

	var setStores map[*coinparam.Params]cxdb.SettlementStore
	if setStores, err = cxdbsql.CreateSettlementStoreMap(coinList); err != nil {
		logging.Fatalf("Error creating settlement store map for opencxd: %s", err)
	}

	// Anyways, here's where we set the server
	var ocxServer *cxserver.OpencxServer
	if ocxServer, err = cxserver.InitServer(setEngines, mengines, limBooks, depositStores, setStores, conf.OpencxHomeDir); err != nil {
		logging.Fatalf("Error initializing server for opencxd: %s", err)
	}

	// For debugging but also it looks nice
	for _, coin := range coinList {
		logging.Infof("Coin supported: %s", coin.Name)
	}

	// Check that the private key exists and if it does, load it
	if err = ocxServer.SetupServerKeys(key); err != nil {
		logging.Fatalf("Error setting up server keys: \n%s", err)
	}

	if conf.LightningSupport {
		// start the lit node for the exchange
		if err = ocxServer.SetupLitNode(key, "lit", "http://hubris.media.mit.edu:46580", "", ""); err != nil {
			logging.Fatalf("Error starting lit node: \n%s", err)
		}

		// register important event handlers -- figure out something better with lightning connection interface
		logging.Infof("registering sigproof handler")
		ocxServer.ExchangeNode.Events.RegisterHandler("qln.chanupdate.sigproof", ocxServer.GetSigProofHandler())
		logging.Infof("done registering sigproof handler")

		logging.Infof("registering opconfirm handler")
		ocxServer.ExchangeNode.Events.RegisterHandler("qln.chanupdate.opconfirm", ocxServer.GetOPConfirmHandler())
		logging.Infof("done registering opconfirm handler")

		logging.Infof("registering push handler")
		ocxServer.ExchangeNode.Events.RegisterHandler("qln.chanupdate.push", ocxServer.GetPushHandler())
		logging.Infof("done registering push handler")

		// Generate the host param list
		// the host params are all of the coinparams / coins we support
		// this coinparam list is generated from the configuration file with generateHostParams
		hpList := util.HostParamList(generateHostParams(&conf))

		// Set up all chain hooks and wallets
		if err = ocxServer.SetupAllWallets(hpList, "wallit/", conf.Resync); err != nil {
			logging.Fatalf("Error setting up wallets: \n%s", err)
			return
		}

		// Waited until the wallets are started, time to link them!
		if err = ocxServer.LinkAllWallets(); err != nil {
			logging.Fatalf("Could not link wallets: \n%s", err)
		}

		// Listen on a bunch of ports according to the number of peers you want to support.
		for portNum := conf.MinPeerPort; portNum < conf.MinPeerPort+conf.MaxPeers; portNum++ {
			var _ string
			if _, err = ocxServer.ExchangeNode.TCPListener(int(portNum)); err != nil {
				return
			}

			// logging.Infof("Listening for connections with address %s on port %d", addr, portNum)
		}

		// Setup lit node rpc
		go ocxServer.SetupLitRPCConnect(conf.Lithost, conf.Litport)

	}

	// Register RPC Commands and set server
	rpc1 := new(cxrpc.OpencxRPC)
	rpc1.OffButton = make(chan bool, 1)
	rpc1.Server = ocxServer

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
			rpc1.OffButton <- true

			return
		}
	}()

	if !conf.AuthenticatedRPC {
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on rpc ===")
		go cxrpc.RPCListenAsync(doneChan, rpc1, conf.Rpchost, conf.Rpcport)
		// block until rpclisten is done
		<-doneChan
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		doneChan := make(chan bool, 1)
		logging.Infof(" === will start to listen on noise-rpc ===")
		go cxrpc.NoiseListenAsync(doneChan, privkey, rpc1, conf.Rpchost, conf.Rpcport)
		// block until noiselisten is done
		<-doneChan

	}

	return
}
