package main

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/mit-dci/lit/coinparam"
	util "github.com/mit-dci/opencx/chainutils"

	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/lit/lnutil"
	litLogging "github.com/mit-dci/lit/logging"
	"github.com/mit-dci/opencx/logging"
)

var (

	// used in init file, so separate
	defaultLogLevel       = 0
	defaultLitLogLevel    = 0
	defaultConfigFilename = "frred.conf"
	defaultLogFilename    = "dblog.txt"
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
	defaultArgs := []byte("tn3=y\ntvtc=y")
	_, err = writer.Write(defaultArgs)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func opencxSetup(conf *frredConfig) *[32]byte {
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
	_, err := os.Stat(conf.FrredHomeDir)
	if err != nil {
		logging.Infof("Creating a home directory at %s", conf.FrredHomeDir)
	}
	if os.IsNotExist(err) {
		os.MkdirAll(conf.FrredHomeDir, 0700)
		logging.Infof("Creating a new config file")
		if err := createDefaultConfigFile(conf.FrredHomeDir); err != nil {
			logging.Fatalf("Error creating a default config file in %v: %s", conf.FrredHomeDir, err)
		}
	}

	if _, err := os.Stat(filepath.Join(conf.FrredHomeDir, defaultConfigFilename)); os.IsNotExist(err) {
		// if there is no config file found over at the directory, create one
		logging.Infof("Creating a new config file")
		err := createDefaultConfigFile(filepath.Join(conf.FrredHomeDir)) // Source of error
		if err != nil {
			logging.Fatal(err)
		}
	}
	conf.ConfigFile = filepath.Join(conf.FrredHomeDir, defaultConfigFilename)
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

	logFilePath := filepath.Join(conf.FrredHomeDir, conf.LogFilename)
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

	keyPath := filepath.Join(conf.FrredHomeDir, defaultKeyFileName)
	privkey, err := lnutil.ReadKeyFile(keyPath)
	if err != nil {
		logging.Fatalf("Error reading key from file: \n%s", err)
	}

	return privkey
}

func generateCoinList(conf *frredConfig) []*coinparam.Params {
	return util.HostParamList(generateHostParams(conf)).CoinListFromHostParams()
}

func generateHostParams(conf *frredConfig) (hostParamList []*util.HostParams) {
	// Regular networks (Just like don't use any of these, I support them though)
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.BitcoinParams, Host: conf.Btchost})
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinParams, Host: conf.Vtchost})
	// Wait until supported by lit
	// hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LitecoinParams, Host: conf.Ltchost})

	// Test nets
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.TestNet3Params, Host: conf.Tn3host})
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinTestNetParams, Host: conf.Tvtchost})
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LiteCoinTestNet4Params, Host: conf.Lt4host})

	// Regression nets
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.RegressionNetParams, Host: conf.Reghost})
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.VertcoinRegTestParams, Host: conf.Rtvtchost})
	hostParamList = append(hostParamList, &util.HostParams{Param: &coinparam.LiteRegNetParams, Host: conf.Litereghost})
	return
}
