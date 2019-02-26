package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/opencx/logging"
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
	defaultArgs := []byte("reg=y\nlitereg=y\nrtvtc=y")
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
		logging.Errorf("Error while creating a directory")
	}
	if os.IsNotExist(err) {
		os.Mkdir(conf.OpencxHomeDir, 0700)
		logging.Infof("Creating a new config file")
		err := createDefaultConfigFile(conf.OpencxHomeDir)
		if err != nil {
			fmt.Printf("Error creating a default config file: %v", conf.OpencxHomeDir)
			logging.Fatal(err)
		}
	}

	if _, err := os.Stat(filepath.Join(conf.OpencxHomeDir, "opencx.conf")); os.IsNotExist(err) {
		// if there is no config file found over at the directory, create one
		logging.Infof("Creating a new config file")
		err := createDefaultConfigFile(filepath.Join(conf.OpencxHomeDir)) // Source of error
		if err != nil {
			logging.Fatal(err)
		}
	}
	conf.ConfigFile = filepath.Join(conf.OpencxHomeDir, "opencx.conf")
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

	logLevel := 0
	if len(conf.LogLevel) == 1 { // -v
		logLevel = 1
	} else if len(conf.LogLevel) == 2 { // -vv
		logLevel = 2
	} else if len(conf.LogLevel) >= 3 { // -vvv
		logLevel = 3
	}
	logging.SetLogLevel(logLevel) // defaults to zero

	keyPath := filepath.Join(conf.OpencxHomeDir, defaultKeyFileName)
	privkey, err := lnutil.ReadKeyFile(keyPath)
	if err != nil {
		logging.Errorf("Error reading key from file: \n%s", err)
	}

	return privkey
}
