package main

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/crypto/koblitz"

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
		return
	}
	return
}

// SignBytes is used in the register method because that's an interactive process.
// BenchClient shouldn't be responsible for interactive stuff, just providing a good
// Go API for the RPC methods the exchange offers.
func (cl *openCxClient) SignBytes(bytes []byte) (signature []byte, err error) {
	var privkeyBytes *[32]byte
	if privkeyBytes, err = lnutil.ReadKeyFile(cl.KeyPath); err != nil {
		logging.Errorf("Error reading key from file: \n%s", err)
		return
	}

	privkey, _ := koblitz.PrivKeyFromBytes(koblitz.S256(), privkeyBytes[:])

	sha := sha3.New256()
	sha.Write(bytes)
	e := sha.Sum(nil)

	if signature, err = koblitz.SignCompact(koblitz.S256(), privkey, e, false); err != nil {
		logging.Errorf("Failed to sign bytes.")
		return
	}

	return
}

// RetreivePublicKey returns the public key if it's been unlocked.
func (cl *openCxClient) RetreivePublicKey() (pubkey *koblitz.PublicKey, err error) {
	_, pubkey = koblitz.PrivKeyFromBytes(koblitz.S256(), cl.RPCClient.PrivKey[:])

	if pubkey == nil {
		err = fmt.Errorf("Private key not unlocked, cannot retreive public key")
		return
	}
	return
}
