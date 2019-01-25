package cxserver

import (
	"fmt"

	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/wallit"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB      *ocxsql.DB
	OpencxRoot    string
	OpencxPort    int
	OpencxPrivkey *hdkeychain.ExtendedKey
	// TODO: Put TLS stuff here
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}

// TODO now that I know how to use this hdkeychain stuff, let's figure out how to create addresses to store

// TODO: remove this method

// NewChildAddress creates a child address, to be assigned to an account, from the current pubkey
func (server *OpencxServer) NewChildAddress() error {
	var err error

	params := coinparam.TestNet3Params
	rootKey, err := hdkeychain.NewMaster([]byte("loljkdjksakdsdasdadjkdsjkdsajsjkdjkasdkjashdakjsdhasjkdhdskjasdh"), &params)
	if err != nil {
		return fmt.Errorf("Error creating new key from string: \n%s", err)
	}

	birthHeight := params.StartHeight
	resync := true
	spvhost := "testnet-seed.bluematt.me"
	path := server.OpencxRoot
	proxyURL := ""

	server.OpencxDB.LogPrintf("Making a new wallet.......\n")
	wallit, _, err := wallit.NewWallit(rootKey, birthHeight, resync, spvhost, path, proxyURL, &params)
	if err != nil {
		return fmt.Errorf("Error when creating wallit: \n%s", err)
	}
	server.OpencxDB.LogPrintf("Made a new wallet??????\n")

	syncHeight, err := wallit.GetDBSyncHeight()
	if err != nil {
		return fmt.Errorf("Error when getting db sync height: \n%s", err)
	}
	server.OpencxDB.LogPrintf("Ayy address: %d\n", syncHeight)

	return nil
}

// SetupServerWallet just loads a private key from a file wallet
func (server *OpencxServer) SetupServerWallet(keypath string) error {
	privkey, err := lnutil.ReadKeyFile(keypath)
	if err != nil {
		return fmt.Errorf("Error reading key from file: \n%s", err)
	}

	rootKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.TestNet3Params)
	if err != nil {
		return fmt.Errorf("Error creating master key from private key: \n%s", err)
	}

	server.OpencxPrivkey = rootKey

	return nil
}
