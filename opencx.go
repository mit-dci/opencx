package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/cxserver"
	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

var (
	logFilename        = "dblog.txt"
	defaultRoot        = ".opencx/"
	defaultPort        = 12345
	defaultKeyFileName = "privkey.hex"
	orderBufferSize    = 1000
)

func main() {
	var err error

	logging.SetLogLevel(2)

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error getting user, needed for directories: \n%s", err)
	}
	defaultRoot = usr.HomeDir + "/" + defaultRoot
	defaultLogPath := defaultRoot + logFilename

	// Create root directory
	createRoot(defaultRoot)

	db := new(ocxsql.DB)

	// Set path where output will be written to for all things database
	err = SetLogPath(defaultLogPath)
	if err != nil {
		log.Fatalf("Error setting logger path: \n%s", err)
	}

	// Check and load config params
	// Start database? That can happen in SetupClient maybe, for DBs that can be started natively in go
	// Check if DB has saved files, if not then start new DB, if so then load old DB
	err = db.SetupClient()
	if err != nil {
		log.Fatalf("Error setting up sql client: \n%s", err)
	}

	// defer the db closing to when we stop
	defer db.DBHandler.Close()

	// Anyways, here's where we set the server
	ocxServer := new(cxserver.OpencxServer)
	ocxServer.OpencxDB = db
	ocxServer.OpencxRoot = defaultRoot
	ocxServer.OpencxPort = defaultPort

	// Check that the private key exists and if it does, load it
	defaultKeyPath := filepath.Join(defaultRoot, defaultKeyFileName)
	if err = ocxServer.SetupServerKeys(defaultKeyPath); err != nil {
		logging.Fatalf("Error setting up server keys: \n%s", err)
	}

	hookErrorChannel := make(chan error, 3)
	// Set up all chain hooks
	go ocxServer.SetupBTCChainhook(hookErrorChannel)
	go ocxServer.SetupLTCChainhook(hookErrorChannel)
	go ocxServer.SetupVTCChainhook(hookErrorChannel)

	// Wait until all hooks are started to do the rest
	for i := 0; i < 3; i++ {
		firstError := <-hookErrorChannel
		if firstError != nil {
			logging.Fatalf("Error when starting hook: \n%s", firstError)
		}
		logging.Infof("Started hook #%d\n", i+1)
	}

	// init the maps for the server
	ocxServer.InitMatchingMaps()

	// Get all the asset pairs then start the matching loop
	assetPairs := match.GenerateAssetPairs()
	for i, pair := range assetPairs {
		go ocxServer.MatchingLoop(pair, orderBufferSize)
		logging.Infof("Pair %d: %s\n", i, pair)
	}
	// Update the addresses -> ONLY uncomment if you switch chains or something. This exchange isn't really meant to be switching between different testnets all the time
	// if err = ocxServer.UpdateAddresses(); err != nil {
	// 	logging.Fatalf("Error updating addresses: \n%s", err)
	// }

	// Register RPC Commands and set server
	rpc1 := new(cxrpc.OpencxRPC)
	rpc1.Server = ocxServer

	err = rpc.Register(rpc1)
	if err != nil {
		log.Fatalf("Error registering RPC Interface:\n%s", err)
	}

	// Start RPC Server
	listener, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", defaultPort))
	fmt.Printf("Running RPC server on %s\n", listener.Addr().String())
	if err != nil {
		log.Fatal("listen error:", err)
	}

	defer listener.Close()
	rpc.Accept(listener)

}

// SetLogPath sets the log path for the database, and tells it to also print to stdout. This should be changed in the future so only verbose clients log to stdout
func SetLogPath(logPath string) error {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(logFile)
	logging.SetLogFile(mw)
	return nil
}

// createRoot exists to make main more readable
func createRoot(rootDir string) {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		logging.Infof("Creating root directory at %s\n", rootDir)
		os.Mkdir(rootDir, os.ModePerm)
	}
}
