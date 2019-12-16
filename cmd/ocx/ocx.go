package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Rjected/lit/btcutil/hdkeychain"
	"github.com/Rjected/lit/coinparam"

	"golang.org/x/crypto/sha3"

	"github.com/Rjected/lit/crypto/koblitz"
	"github.com/Rjected/lit/portxo"

	"github.com/mit-dci/opencx/benchclient"

	"github.com/Rjected/lit/lnutil"
	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/opencx/logging"
)

type ocxClient struct {
	KeyPath     string
	KeyPassword string
	RPCClient   *benchclient.BenchClient
	unlocked    bool
}

type ocxConfig struct {
	// Filename of config file where this stuff can be set as well
	ConfigFile string

	// stuff for files and directories
	LogFilename string `long:"logFilename" short:"l" description:"Filename for output log file"`
	OcxHomeDir  string `long:"dir" short:"d" description:"Location of the root directory relative to home directory"`

	// stuff for ports
	Rpchost string `long:"rpchost" short:"h" description:"Hostname of OpenCX Server you'd like to connect to"`
	Rpcport uint16 `long:"rpcport" short:"p" description:"Port of the OpenCX Server you'd like to connect to"`

	// filename for key
	KeyFileName string `long:"keyfilename" short:"k" description:"Filename for private key within root opencx directory used to send transactions"`
	// password for the encrypted key file
	// NOTE: This is NOT SECURE! It saves the password in a string and strings
	// are very difficult to zero out since they are immutable, so expect this
	// to remain in memory after being garbage collected.
	// TODO: Figure out a way to have secure non-interactive key encryption /
	// decryption - maybe use memguard and allow things to be piped in.
	KeyPassword string `long:"keypass" description:"Password for encrypted private key file"`

	// logging and debug parameters
	LogLevel []bool `short:"v" description:"Set verbosity level to verbose (-v), very verbose (-vv) or very very verbose (-vvv)"`

	// auth or unauth rpc?
	AuthenticatedRPC bool `long:"authrpc" description:"Whether or not to use authenticated RPC"`
}

// Let these be turned into config things at some point
var (
	defaultConfigFilename   = "ocx.conf"
	defaultLogFilename      = "ocxlog.txt"
	defaultOcxHomeDirName   = os.Getenv("HOME") + "/.opencx/ocx/"
	defaultKeyFileName      = defaultOcxHomeDirName + "privkey.hex"
	defaultLogLevel         = 0
	defaultHomeDir          = os.Getenv("HOME")
	defaultRpcport          = uint16(12345)
	defaultRpchost          = "hubris.media.mit.edu"
	defaultAuthenticatedRPC = true
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *ocxConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

// opencx-cli is the client, opencx is the server
func main() {
	var err error
	var client ocxClient

	conf := &ocxConfig{
		OcxHomeDir:       defaultOcxHomeDirName,
		Rpchost:          defaultRpchost,
		Rpcport:          defaultRpcport,
		LogFilename:      defaultLogFilename,
		KeyFileName:      defaultKeyFileName,
		ConfigFile:       defaultConfigFilename,
		AuthenticatedRPC: defaultAuthenticatedRPC,
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

	if len(conf.KeyPassword) > 0 {
		// TODO: move the password and key loading to this function rather than
		// UnlockKey
		client.KeyPassword = conf.KeyPassword
	}

	client.KeyPath = filepath.Join(conf.KeyFileName)
	client.RPCClient = new(benchclient.BenchClient)
	if !conf.AuthenticatedRPC {
		if err = client.RPCClient.SetupBenchClient(conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
		}
	} else {
		// We have to unlock the key twice, once for handshake and again for commands
		if err = client.UnlockKey(); err != nil {
			return
		}
		if err = client.RPCClient.SetupBenchNoiseClient(conf.Rpchost, conf.Rpcport); err != nil {
			logging.Fatalf("Error setting up OpenCX RPC Client: \n%s", err)
		}

	}

	if err = client.parseCommands(os.Args[1:]); err != nil {
		logging.Fatalf("%s", err)
	}
}

func (cl *ocxClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return cl.RPCClient.Call(serviceMethod, args, reply)
}

func (cl *ocxClient) UnlockKey() (err error) {
	// if we're not unlocked and the client is fine too then don't bother
	if !cl.unlocked || cl.RPCClient.PrivKey == nil {
		var keyFromFile *[32]byte
		logging.Infof("Client keypath: %s", cl.KeyPath)

		// START OF SURGERY
		if len(cl.KeyPassword) > 0 {
			var zeroArray []byte = make([]byte, len(cl.KeyPassword))
			var keyPasswordBytes []byte = make([]byte, len(cl.KeyPassword))
			copy(keyPasswordBytes, []byte(cl.KeyPassword))

			zero32 := [32]byte{}
			key32 := new([32]byte)
			_, err = os.Stat(cl.KeyPath)
			if err != nil {
				if os.IsNotExist(err) {
					// no key found, generate and save one
					logging.Infof("No file %s, generating.\n", cl.KeyPath)

					_, err = rand.Read(key32[:])
					if err != nil {
						logging.Errorf("Error reading from rand into key: %s", err)
						return
					}

					err = lnutil.SaveKeyToFileArg(cl.KeyPath, key32, keyPasswordBytes)
					if err != nil {
						logging.Errorf("Error saving key to file: %s", err)
						return
					}
				} else {
					// unknown error, crash
					err = fmt.Errorf("Unknown error reading keyfile")
					logging.Errorf("Client UnlockKey Error: %s", err)
					return
				}
			}
			// zero it out
			copy(key32[:], zero32[:])

			// now load from file
			keyFromFile, err = lnutil.LoadKeyFromFileArg(cl.KeyPath, keyPasswordBytes)
			if err != nil {
				logging.Errorf("Error reading key from file with password arg: \n%s", err)
				return
			}

			// Zero array - the string isn't going to be zeroed but this array will,
			// so at least the password is in one less place
			copy(keyPasswordBytes, zeroArray)
		} else {
			keyFromFile, err = lnutil.ReadKeyFile(cl.KeyPath)
			if err != nil {
				logging.Errorf("Error reading key from file: \n%s", err)
				return
			}
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
		cl.unlocked = true
	} else {
		logging.Infof("Using already unlocked key")
	}
	// logging.Infof("Public Key (compressed): %x", cl.RPCClient.PrivKey.PubKey().SerializeCompressed())

	return
}

// SignBytes is used in the register method because that's an interactive process.
// BenchClient shouldn't be responsible for interactive stuff, just providing a good
// Go API for the RPC methods the exchange offers.
func (cl *ocxClient) SignBytes(bytes []byte) (signature []byte, err error) {

	sha := sha3.New256()
	sha.Write(bytes)
	e := sha.Sum(nil)

	if signature, err = koblitz.SignCompact(koblitz.S256(), cl.RPCClient.PrivKey, e, false); err != nil {
		logging.Errorf("Failed to sign bytes.")
		return
	}

	return
}

// RetrievePublicKey returns the public key if it's been unlocked.
func (cl *ocxClient) RetrievePublicKey() (pubkey *koblitz.PublicKey, err error) {
	if !cl.unlocked {
		err = fmt.Errorf("Key not unlocked, cannot retrieve pubkey")
		return
	}
	pubkey = cl.RPCClient.PrivKey.PubKey()
	return
}
