package ocxredis

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mit-dci/opencx/logging"

	"github.com/go-redis/redis"
)

// Let these be turned into config things at some point
var (
	defaultServer = "localhost"
	defaultPort   = 6379
)

// DB holds the client for the redis DB
type DB struct {
	dbClient *redis.Client
	logger   *log.Logger
}

// SetupClient sets up the redis client
func (db *DB) SetupClient() error {

	db.dbClient = redis.NewClient(&redis.Options{
		Addr:     defaultServer + ":" + fmt.Sprintf("%d", defaultPort),
		Password: "",
		DB:       0,
	})

	// Check that the database is working / running
	status := db.dbClient.Ping()
	if status.Err() != nil {
		return fmt.Errorf("Error when pinging redis server, is your database running?: \n%s", status.Err())
	}
	return nil
}

// SetDataDirectory sets the data directory for the redis client
func (db *DB) SetDataDirectory(dataDirectory string) error {
	// Create dataDirectory if it doesn't exist
	if _, err := os.Stat(dataDirectory); os.IsNotExist(err) {
		os.Mkdir(dataDirectory, os.ModePerm)
	}

	status := db.dbClient.ConfigSet("dir", dataDirectory)
	if status.Err() != nil {
		return fmt.Errorf("Error when setting data directory for database: \n%s", status.Err())
	}
	return nil
}

// SetLogPath sets the log path for the database, and tells it to also print to stdout. This should be changed in the future so only verbose clients log to stdout
func (db *DB) SetLogPath(logPath string) error {
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	logging.SetLogFile(mw)
	return nil
}
