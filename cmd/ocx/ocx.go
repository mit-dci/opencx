package main

import (
	"os"
	"path/filepath"

	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"

	"golang.org/x/crypto/sha3"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/lit/portxo"

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
	// Filename of config file where this stuff can be set as well
	ConfigFile string

	// stuff for files and directories
	LogFilename string `long:"logFilename" short:"l" description:"Filename for output log file"`
	OcxHomeDir  string `long:"dir" short:"d" description:"Location of the root directory relative to home directory"`

	// stuff for ports
	Rpchost string `long:"rpchost" short:"h" description:"Hostname of OpenCX Server you'd like to connect to"`
	Rpcport uint16 `long:"rpcport" short:"p" description:"Port of the OpenCX Port you'd like to connect to"`

	// filename for key
	KeyFileName string `long:"keyfilename" short:"k" description:"Filename for private key within root opencx directory used to send transactions"`

	// logging and debug parameters
	LogLevel []bool `short:"v" description:"Set verbosity level to verbose (-v), very verbose (-vv) or very very verbose (-vvv)"`
}

// Let these be turned into config things at some point
var (
	defaultConfigFilename = "ocx.conf"
	defaultLogFilename    = "ocxlog.txt"
	defaultOcxHomeDirName = os.Getenv("HOME") + "/.ocx/"
	defaultKeyFileName    = defaultOcxHomeDirName + "privkey.hex"
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
		OcxHomeDir:  defaultOcxHomeDirName,
		Rpchost:     defaultRpchost,
		Rpcport:     defaultRpcport,
		LogFilename: defaultLogFilename,
		KeyFileName: defaultKeyFileName,
		ConfigFile:  defaultConfigFilename,
	}

	if didWriteHelp := ocxSetup(conf); didWriteHelp {
		return
	}

	if len(os.Args) < 2 {
		logging.Fatalf("Please enter arguments to the command line tool")
		return
	}

	if os.Args[1] == "help" {
		if err = client.parseCommands(os.Args[1:]); err != nil {
			logging.Fatalf("%s", err)
		}
		return
	}
	client.KeyPath = filepath.Join(conf.KeyFileName)
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
	var keyFromFile *[32]byte
	logging.Infof("client keypath: %s", cl.KeyPath)
	if keyFromFile, err = lnutil.ReadKeyFile(cl.KeyPath); err != nil {
		logging.Errorf("Error reading key from file: \n%s", err)
		return
	}

	// We use TestNet3Params because that's what qln uses
	var rootPrivKey *hdkeychain.ExtendedKey
	if rootPrivKey, err = hdkeychain.NewMaster(keyFromFile[:], &coinparam.TestNet3Params); err != nil {
		return
	}

	// make keygen the same
	var kg portxo.KeyGen
	kg.Depth = 5
	kg.Step[0] = 44 | 1<<31
	kg.Step[1] = 513 | 1<<31
	kg.Step[2] = 9 | 1<<31
	kg.Step[3] = 0 | 1<<31
	kg.Step[4] = 0 | 1<<31
	if cl.RPCClient.PrivKey, err = kg.DerivePrivateKey(rootPrivKey); err != nil {
		return
	}

	return
}

// SignBytes is used in the register method because that's an interactive process.
// BenchClient shouldn't be responsible for interactive stuff, just providing a good
// Go API for the RPC methods the exchange offers.
func (cl *openCxClient) SignBytes(bytes []byte) (signature []byte, err error) {

	sha := sha3.New256()
	sha.Write(bytes)
	e := sha.Sum(nil)

	if signature, err = koblitz.SignCompact(koblitz.S256(), cl.RPCClient.PrivKey, e, false); err != nil {
		logging.Errorf("Failed to sign bytes.")
		return
	}

	return
}

// RetreivePublicKey returns the public key if it's been unlocked.
func (cl *openCxClient) RetreivePublicKey() (pubkey *koblitz.PublicKey) {
	pubkey = cl.RPCClient.PrivKey.PubKey()
	return
}
