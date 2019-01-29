package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mit-dci/opencx/cxserver"

	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/lit/lnutil"
)

var (
	logFilename        = "dblog.txt"
	defaultRoot        = ".opencx/"
	defaultPort        = 12345
	defaultKeyFileName = "privkey.hex"
)

func main() {
	var err error

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
	err = db.SetLogPath(defaultLogPath)
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

	// Anyways, here's where we set the server
	ocxServer := new(cxserver.OpencxServer)
	ocxServer.OpencxDB = db
	ocxServer.OpencxRoot = defaultRoot
	ocxServer.OpencxPort = defaultPort

	// Check that the private key exists and if it does, load it
	defaultKeyPath := filepath.Join(defaultRoot, defaultKeyFileName)
	ocxServer.SetupServerKeys(defaultKeyPath)

	heightEventChannel := ocxServer.LetMeKnowHeight()
	go heightEventHandler(heightEventChannel)
	err = ocxServer.SetupBTCChainhook()
	if err != nil {
		log.Fatalf("Error setting up btc chainhook:\n%s", err)
	}

	// ocxServer.SetupWallets()

	// defer the db to when it closes
	defer ocxServer.OpencxDB.DBHandler.Close()

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

	// TODO: do TLS stuff here so its secure

	defer listener.Close()
	rpc.Accept(listener)

}

// createRoot exists to make main more readable
func createRoot(rootDir string) {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		fmt.Printf("Creating root directory at %s\n", rootDir)
		os.Mkdir(rootDir, os.ModePerm)
	}
}

// TODO: Clean up the rest of the code then figure out if this should be removed
func heightEventHandler(HeightEventChan chan lnutil.HeightEvent) {
	for {
		_ = <- HeightEventChan
		// Why does this work
	}
}
