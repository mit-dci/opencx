package main

import (
	"encoding/hex"
	"os"
	"os/signal"
	"syscall"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/crypto/koblitz"

	flags "github.com/jessevdk/go-flags"
	util "github.com/mit-dci/opencx/chainutils"
	"github.com/mit-dci/opencx/cxdb"
	"github.com/mit-dci/opencx/cxdb/cxdbmemory"
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
	MaxPeers    uint16   `long:"numpeers" description:"Maximum number of peers that you'd like to support"`
	MinPeerPort uint16   `long:"minpeerport" description:"Port to start creating ports for peers at"`
	Lithost     string   `long:"lithost" description:"Host for the lightning node on the exchange to run"`
	Litport     uint16   `long:"litport" description:"Port for the lightning node on the exchange to run"`
	Whitelist   []string `long:"whitelist" description:"If using pinky swear settlement, this is the default whitelist"`

	// filename for key
	KeyFileName string `long:"keyfilename" short:"k" description:"Filename for private key within root opencx directory used to send transactions"`
	// password for the encrypted key file
	// NOTE: This is NOT SECURE! It saves the password in a string and strings
	// are very difficult to zero out since they are immutable, so expect this
	// to remain in memory after being garbage collected.
	// TODO: Figure out a way to have secure non-interactive key encryption /
	// decryption - maybe use memguard and allow things to be piped in.
	KeyPassword string `long:"keypass" description:"Password for encrypted private key file"`

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

	logging.Infof("Creating limit engines...")
	var mengines map[match.Pair]match.LimitEngine
	if mengines, err = cxdbsql.CreateLimitEngineMap(pairList); err != nil {
		logging.Fatalf("Error creating limit engine map with coinlist for opencxd: %s", err)
	}

	var setEngines map[*coinparam.Params]match.SettlementEngine
	if len(conf.Whitelist) != 0 {
		whitelist := make([][33]byte, len(conf.Whitelist))
		var pkBytes []byte
		for i, str := range conf.Whitelist {
			if pkBytes, err = hex.DecodeString(str); err != nil {
				logging.Fatalf("Error decoding string for whitelist: %s", err)
			}
			if len(pkBytes) != 33 {
				logging.Fatalf("One pubkey not 33 bytes")
			}
			logging.Infof("Adding %x to the whitelist", pkBytes)
			copy(whitelist[i][:], pkBytes)
		}
		whitelistMap := make(map[*coinparam.Params][][33]byte)
		for _, coin := range coinList {
			whitelistMap[coin] = whitelist
		}
		logging.Infof("Creating pinky swear engines...")
		if setEngines, err = cxdbmemory.CreatePinkySwearEngineMap(whitelistMap, true); err != nil {
			logging.Fatalf("Error creating pinky swear settlement engine map for opencxd: %s", err)
		}
	} else {
		logging.Infof("Creating settlement engines...")
		if setEngines, err = cxdbsql.CreateSettlementEngineMap(coinList); err != nil {
			logging.Fatalf("Error creating settlement engine map for opencxd: %s", err)
		}
	}
	if setEngines == nil {
		logging.Fatalf("Error, nil setEngines map, this should not ever happen")
	}

	logging.Infof("Creating limit orderbooks...")
	var limBooks map[match.Pair]match.LimitOrderbook
	if limBooks, err = cxdbsql.CreateLimitOrderbookMap(pairList); err != nil {
		logging.Fatalf("Error creating limit orderbook map for opencxd: %s", err)
	}

	logging.Infof("Creating deposit stores...")
	var depositStores map[*coinparam.Params]cxdb.DepositStore
	if depositStores, err = cxdbsql.CreateDepositStoreMap(coinList); err != nil {
		logging.Fatalf("Error creating deposit store map for opencxd: %s", err)
	}

	logging.Infof("Creating settlement stores...")
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

	// Generate the host param list
	// the host params are all of the coinparams / coins we support
	// this coinparam list is generated from the configuration file with generateHostParams
	hpList := util.HostParamList(generateHostParams(&conf))

	// Set up all chain hooks and wallets
	if err = ocxServer.SetupAllWallets(hpList, "wallit/", conf.Resync); err != nil {
		logging.Fatalf("Error setting up wallets: \n%s", err)
		return
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

	var rpcListener *cxrpc.OpencxRPCCaller
	if rpcListener, err = cxrpc.CreateRPCForServer(ocxServer); err != nil {
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

			// stop rpc listener
			if err = rpcListener.Stop(); err != nil {
				logging.Fatalf("Error killing server: %s", err)
			}

			return
		}
	}()

	if !conf.AuthenticatedRPC {
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on rpc ===")
		if err = rpcListener.RPCListen(conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error listening for rpc for server: %s", err)
		}
	} else {

		privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), key[:])
		// this tells us when the rpclisten is done
		logging.Infof(" === will start to listen on noise-rpc ===")
		if err = rpcListener.NoiseListen(privkey, conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error listening for noise rpc for server: %s", err)
		}

	}

	// wait until the listener dies - this does not return anything
	rpcListener.WaitUntilDead()

	return
}
