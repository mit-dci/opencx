package main

import (
	"os"
	"path/filepath"

	"github.com/mit-dci/opencx/cmd/benchclient"

	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/logging"
)

type openCxClient struct {
	KeyPath   string
	RPCClient *benchclient.BenchClient
}

type ocxConfig struct {
	ConfigFile string

	// stuff for files and directories
	LogFilename string `long:"logFilename" description:"Filename for output log file"`
	OcxHomeDir  string `long:"dir" description:"Location of the root directory relative to home directory"`

	// stuff for ports
	Rpchost string `long:"rpchost" description:"Hostname of OpenCX Server you'd like to connect to"`
	Rpcport uint16 `long:"rpcport" description:"Port of the OpenCX Port you'd like to connect to"`

	// logging and debug parameters
	LogLevel []bool `short:"v" description:"Set verbosity level to verbose (-v), very verbose (-vv) or very very verbose (-vvv)"`
}

// Let these be turned into config things at some point
var (
	defaultConfigFilename = "ocx.conf"
	defaultLogFilename    = "ocxlog.txt"
	defaultOcxHomeDirName = os.Getenv("HOME") + "/.ocx/"
	defaultKeyFileName    = "privkey.hex"
	defaultLogLevel       = 0
	defaultHomeDir        = os.Getenv("HOME")
	defaultRpcport        = uint16(12345)
	defaultRpchost        = "hubris.media.mit.edu"
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *ocxConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

// opencx-cli is the client, opencx is the server
func main() {
	var err error
	var client openCxClient

	conf := &ocxConfig{
		OcxHomeDir: defaultOcxHomeDirName,
		Rpchost:    defaultRpchost,
		Rpcport:    defaultRpcport,
	}

	ocxSetup(conf)
	client.KeyPath = filepath.Join(conf.OcxHomeDir, defaultKeyFileName)
	client.RPCClient = new(benchclient.BenchClient)
	if err = client.RPCClient.SetupBenchClient(conf.Rpchost, conf.Rpcport); err != nil {
		logging.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
	}

	if err = client.parseCommands(os.Args[1:]); err != nil {
		logging.Fatalf("%s", err)
	}
}

func (cl *openCxClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}

func (cl *openCxClient) UnlockKey() (err error) {
	if cl.RPCClient.PrivKey, err = lnutil.ReadKeyFile(cl.KeyPath); err != nil {
		logging.Errorf("Error reading key from file: \n%s", err)
	}
	return
}
