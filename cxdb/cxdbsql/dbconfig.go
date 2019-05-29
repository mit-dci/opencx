package cxdbsql

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
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
	defaultArgs := []byte("\n")
	_, err = writer.Write(defaultArgs)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func dbConfigSetup(conf *dbsqlConfig) {
	// Pre-parse the command line options to see if an alternative config
	// file or the version flag was specified. Config file will be read later
	// and cli options would be parsed again below

	parser := newConfigParser(conf, flags.Default)

	// create home directory
	_, err := os.Stat(conf.DBHomeDir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(conf.DBHomeDir, 0700); err != nil {
			logging.Fatalf("Could not make dirs needed for home dir %s", conf.DBHomeDir)
		}

		logging.Infof("Creating a new db home directory at %s", conf.DBHomeDir)

		if err = createDefaultConfigFile(conf.DBHomeDir); err != nil {
			logging.Fatalf("Error creating a default config file: %s", conf.DBHomeDir)
		}
	} else if err != nil {
		logging.Fatalf("Error while creating a directory: %s", err)
	}

	if _, err := os.Stat(filepath.Join(conf.DBHomeDir, defaultConfigFilename)); os.IsNotExist(err) {
		// if there is no config file found over at the directory, create one
		logging.Infof("Creating a new default db config file at %s", conf.DBHomeDir)

		// source of error
		if err = createDefaultConfigFile(filepath.Join(conf.DBHomeDir)); err != nil {
			logging.Fatal(err)
		}
	}
	conf.ConfigFile = filepath.Join(conf.DBHomeDir, defaultConfigFilename)
	// lets parse the config file provided, if any
	if err = flags.NewIniParser(parser).ParseFile(conf.ConfigFile); err != nil {
		// If the error isn't a path error then we care about it
		if _, ok := err.(*os.PathError); !ok {
			logging.Fatalf("Non-path error encountered when parsing config file: %s", err)
		}
	}

	return
}
