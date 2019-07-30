package cxdbsql

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/mit-dci/opencx/logging"
)

type dbsqlConfig struct {
	// Filename of config file where this stuff can be set as well
	ConfigFile string

	// database home dir
	DBHomeDir string `long:"dir" description:"Location of the root directory for the sql db info and config"`

	// database info required to establish connection
	DBUsername string `long:"dbuser" description:"database username"`
	DBPassword string `long:"dbpassword" description:"database password"`
	DBHost     string `long:"dbhost" description:"Host for the database connection"`
	DBPort     uint16 `long:"dbport" description:"Port for the database connection"`

	// database schema names
	ReadOnlyOrderSchemaName   string `long:"readonlyorderschema" description:"Name of read-only orderbook schema"`
	ReadOnlyAuctionSchemaName string `long:"readonlyauctionschema" description:"Name of read-only auction schema"`
	ReadOnlyBalanceSchemaName string `long:"readonlybalanceschema" description:"Name of read-only balance schema"`
	BalanceSchemaName         string `long:"balanceschema" description:"Name of balance schema"`
	DepositSchemaName         string `long:"depositschema" description:"Name of deposit schema"`
	PendingDepositSchemaName  string `long:"penddepschema" description:"Name of pending deposit schema"`
	PuzzleSchemaName          string `long:"puzzleschema" description:"Name of schema for puzzle orderbooks"`
	AuctionSchemaName         string `long:"auctionschema" description:"Name of schema for auction ID"`
	AuctionOrderSchemaName    string `long:"auctionorderschema" description:"Name of schema for auction orderbook"`
	OrderSchemaName           string `long:"orderschema" description:"Name of schema for limit orderbook"`
	PeerSchemaName            string `long:"peerschema" description:"Name of schema for peer storage"`

	// database table names
	PuzzleTableName       string `long:"puzzletable" description:"Name of table for puzzle orderbooks"`
	AuctionOrderTableName string `long:"auctionordertable" description:"Name of table for auction orders"`
	PeerTableName         string `long:"peertable" description:"Name of table for peer storage"`
}

// Let these be turned into config things at some point
var (
	defaultConfigFilename = "sqldb.conf"
	defaultHomeDir        = os.Getenv("HOME")
	defaultDBHomeDirName  = defaultHomeDir + "/.opencx/db/"
	defaultDBPort         = uint16(3306)
	defaultDBHost         = "localhost"
	defaultDBUser         = "opencx"
	defaultDBPass         = "testpass"

	// definitely move this to a config file
	defaultReadOnlyOrderSchema   = "orders_readonly"
	defaultReadOnlyAuctionSchema = "auctionorders_readonly"
	defaultReadOnlyBalanceSchema = "balances_readonly"
	defaultBalanceSchema         = "balances"
	defaultDepositSchema         = "deposit"
	defaultPendingDepositSchema  = "pending_deposits"
	defaultPuzzleSchema          = "puzzle"
	defaultAuctionSchema         = "auctions"
	defaultAuctionOrderSchema    = "auctionorder"
	defaultOrderSchema           = "orders"
	defaultPeerSchema            = "peers"

	// tables
	defaultAuctionOrderTable = "auctionorders"
	defaultPuzzleTable       = "puzzles"
	defaultPeerTable         = "opencxpeers"

	// Set defaults
	defaultConf = &dbsqlConfig{
		// home dir
		DBHomeDir: defaultDBHomeDirName,

		// user / pass / net stuff
		DBUsername: defaultDBUser,
		DBPassword: defaultDBPass,
		DBHost:     defaultDBHost,
		DBPort:     defaultDBPort,

		// schemas
		ReadOnlyAuctionSchemaName: defaultReadOnlyAuctionSchema,
		ReadOnlyOrderSchemaName:   defaultReadOnlyOrderSchema,
		ReadOnlyBalanceSchemaName: defaultReadOnlyBalanceSchema,
		BalanceSchemaName:         defaultBalanceSchema,
		DepositSchemaName:         defaultDepositSchema,
		PendingDepositSchemaName:  defaultPendingDepositSchema,
		PuzzleSchemaName:          defaultPuzzleSchema,
		AuctionSchemaName:         defaultAuctionSchema,
		AuctionOrderSchemaName:    defaultAuctionOrderSchema,
		OrderSchemaName:           defaultOrderSchema,
		PeerSchemaName:            defaultPeerSchema,

		// tables
		PuzzleTableName:       defaultPuzzleTable,
		AuctionOrderTableName: defaultAuctionOrderTable,
		PeerTableName:         defaultPeerTable,
	}
)

// newConfigParser returns a new command line flags parser.
func newConfigParser(conf *dbsqlConfig, options flags.Options) *flags.Parser {
	parser := flags.NewParser(conf, options)
	return parser
}

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
	defaultArgs := []byte("dbuser=opencx\ndbpassword=testpass\n")
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
