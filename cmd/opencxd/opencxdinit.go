package main

import (
	"bufio"
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/Rjected/lit/coinparam"
	util "github.com/mit-dci/opencx/chainutils"

	"github.com/Rjected/lit/lnutil"
	litLogging "github.com/Rjected/lit/logging"
	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/opencx/logging"
)

var (

	// used in init file, so separate
	defaultLogLevel       = 0
	defaultLitLogLevel    = 0
	defaultConfigFilename = "opencx.conf"
	defaultLogFilename    = "opencxdlog.txt"
	defaultKeyFileName    = "privkey.hex"
)

// createDefaultConfigFile creates a config file  -- only call this if the
// config file isn't already there
func createDefaultConfigFile(destinationPath string) error {

	dest, err := os.OpenFile(filepath.Join(destinationPath, defaultConfigFilename),
		os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer dest.Close()

	writer := bufio.NewWriter(dest)
	defaultArgs := []byte("tn3=y\ntvtc=y\n")
	_, err = writer.Write(defaultArgs)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func opencxSetup(conf *opencxConfig) *[32]byte {
	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified. Config file will be read later
	// and cli options would be parsed again below

	parser := newConfigParser(conf, flags.Default)

	if _, err := parser.ParseArgs(os.Args); err != nil {
		// catch all cli argument errors
		logging.Fatal(err)
	}

	// set default log level
	logging.SetLogLevel(defaultLogLevel)

	// create home directory
	_, err := os.Stat(conf.OpencxHomeDir)
	if err != nil {
		logging.Infof("Creating a home directory at %s", conf.OpencxHomeDir)
	}
	if os.IsNotExist(err) {
		os.MkdirAll(conf.OpencxHomeDir, 0700)
		logging.Infof("Creating a new config file")
		if err := createDefaultConfigFile(conf.OpencxHomeDir); err != nil {
			logging.Fatalf("Error creating a default config file in %v: %s", conf.OpencxHomeDir, err)
		}
	}

	if _, err := os.Stat(filepath.Join(conf.OpencxHomeDir, defaultConfigFilename)); os.IsNotExist(err) {
		// if there is no config file found over at the directory, create one
		logging.Infof("Creating a new config file")
		err := createDefaultConfigFile(filepath.Join(conf.OpencxHomeDir)) // Source of error
		if err != nil {
			logging.Fatal(err)
		}
	}
	conf.ConfigFile = filepath.Join(conf.OpencxHomeDir, defaultConfigFilename)
	// lets parse the config file provided, if any
	err = flags.NewIniParser(parser).ParseFile(conf.ConfigFile)
	if err != nil {
		_, ok := err.(*os.PathError)
		if !ok {
			logging.Fatal(err)
		}
	}

	// Parse command line options again to ensure they take precedence.
	_, err = parser.ParseArgs(os.Args) // returns invalid flags
	if err != nil {
		logging.Fatal(err)
	}

	logFilePath := filepath.Join(conf.OpencxHomeDir, conf.LogFilename)
	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer logFile.Close()
	logging.SetLogFile(logFile)

	logLevel := defaultLogLevel
	if len(conf.LogLevel) == 1 { // -v
		logLevel = 1
	} else if len(conf.LogLevel) == 2 { // -vv
		logLevel = 2
	} else if len(conf.LogLevel) >= 3 { // -vvv
		logLevel = 3
	}
	logging.SetLogLevel(logLevel) // defaults to defaultLogLevel

	litLogLevel := defaultLitLogLevel
	if len(conf.LitLogLevel) == 1 { // -w
		litLogLevel = 1
	} else if len(conf.LitLogLevel) == 2 { // -ww
		litLogLevel = 2
	} else if len(conf.LitLogLevel) >= 3 { // -www
		litLogLevel = 3
	}
	litLogging.SetLogLevel(litLogLevel) // defaults to defaultLitLogLevel

	keyPath := filepath.Join(conf.OpencxHomeDir, defaultKeyFileName)
	var privkey *[32]byte
	if len(conf.KeyPassword) > 0 {
		var zeroArray []byte = make([]byte, len(conf.KeyPassword))
		var keyPasswordBytes []byte = make([]byte, len(conf.KeyPassword))
		copy(keyPasswordBytes, []byte(conf.KeyPassword))

		zero32 := [32]byte{}
		key32 := new([32]byte)
		_, err := os.Stat(keyPath)
		if err != nil {
			if os.IsNotExist(err) {
				// no key found, generate and save one
				logging.Infof("No file %s, generating.\n", keyPath)

				_, err := rand.Read(key32[:])
				if err != nil {
					logging.Fatalf("Error reading from rand into key: %s", err)
				}

				err = lnutil.SaveKeyToFileArg(keyPath, key32, keyPasswordBytes)
				if err != nil {
					logging.Fatalf("Error saving key to file: %s", err)
				}
			} else {
				// unknown error, crash
				logging.Fatalf("unknown error reading keyfile\n")
			}
		}
		// zero it out
		copy(key32[:], zero32[:])

		// now load from file
		privkey, err = lnutil.LoadKeyFromFileArg(keyPath, keyPasswordBytes)
		if err != nil {
			logging.Fatalf("Error reading key from file with password arg: \n%s", err)
		}

		// Zero array - the string isn't going to be zeroed but this array will,
		// so at least the password is in one less place
		copy(keyPasswordBytes, zeroArray)
	} else {
		privkey, err = lnutil.ReadKeyFile(keyPath)
		if err != nil {
			logging.Fatalf("Error reading key from file: \n%s", err)
		}
	}

	return privkey
}

func generateCoinList(conf *opencxConfig) []*coinparam.Params {
	return util.HostParamList(generateHostParams(conf)).CoinListFromHostParams()
}

func generateHostParams(conf *opencxConfig) (hostParamList []*util.HostParams) {
	// Regular networks (Just like don't use any of these, I support them though)
	if conf.Btchost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.BitcoinParams, Host: conf.Btchost})
	}

	if conf.Vtchost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinParams, Host: conf.Vtchost})
	}
	// Wait until supported by lit
	// hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LitecoinParams, Host: conf.Ltchost})

	// Test nets
	if conf.Tn3host != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.TestNet3Params, Host: conf.Tn3host})
	}

	if conf.Tvtchost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinTestNetParams, Host: conf.Tvtchost})
	}

	if conf.Lt4host != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LiteCoinTestNet4Params, Host: conf.Lt4host})
	}

	// Regression nets
	if conf.Reghost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.RegressionNetParams, Host: conf.Reghost})
	}

	if conf.Rtvtchost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinRegTestParams, Host: conf.Rtvtchost})
	}

	if conf.Litereghost != "" {
		hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LiteRegNetParams, Host: conf.Litereghost})
	}
	return
}
