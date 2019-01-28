package cxserver

import (
	"fmt"
	"os"

	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/uspv"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/lit/wallit"
	"github.com/mit-dci/lit/btcutil/base58"
)

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB             *ocxsql.DB
	OpencxRoot           string
	OpencxPort           int
	// Hehe it's the vault, pls don't steal
	OpencxBTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxVTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxLTCTestPrivKey *hdkeychain.ExtendedKey
	BTCWallet            *wallit.Wallit
	LTCWallet            *wallit.Wallit
	VTCWallet            *wallit.Wallit
	// TODO: Put TLS stuff here
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}

// TODO now that I know how to use this hdkeychain stuff, let's figure out how to create addresses to store

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(keypath string) error {
	privkey, err := lnutil.ReadKeyFile(keypath)
	if err != nil {
		return fmt.Errorf("Error reading key from file: \n%s", err)
	}

	rootBTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.TestNet3Params)
	if err != nil {
		return fmt.Errorf("Error creating master BTC Test key from private key: \n%s", err)
	}

	server.OpencxBTCTestPrivKey = rootBTCKey

	rootVTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.VertcoinRegTestParams)
	if err != nil {
		return fmt.Errorf("Error creating master VTC Test key from private key: \n%s", err)
	}

	server.OpencxVTCTestPrivKey = rootVTCKey

	rootLTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.LiteCoinTestNet4Params)
	if err != nil {
		return fmt.Errorf("Error creating master LTC Test key from private key: \n%s", err)
	}

	server.OpencxLTCTestPrivKey = rootLTCKey

	return nil
}

// SetupWallets ...
func (server *OpencxServer) SetupWallets() error {
	var err error

	btcRoot := server.createSubRoot(coinparam.TestNet3Params.Name)
	ltcRoot := server.createSubRoot(coinparam.LiteCoinTestNet4Params.Name)
	vtcRoot := server.createSubRoot(coinparam.VertcoinTestNetParams.Name)

	btcwallet, _, err := wallit.NewWallit(server.OpencxBTCTestPrivKey, coinparam.TestNet3Params.StartHeight, true, "1", btcRoot, "", &coinparam.TestNet3Params)
	if err != nil {
		return fmt.Errorf("Error setting up wallet: \n%s", err)
	}

	vtcwallet, _, err := wallit.NewWallit(server.OpencxVTCTestPrivKey, coinparam.VertcoinTestNetParams.StartHeight, true, "1", vtcRoot, "", &coinparam.VertcoinTestNetParams)
	if err != nil {
		return fmt.Errorf("Error setting up vtc wallet: \n%s", err)
	}

	ltcwallet, _, err := wallit.NewWallit(server.OpencxLTCTestPrivKey, coinparam.LiteCoinTestNet4Params.StartHeight, true, "1", ltcRoot, "", &coinparam.LiteCoinTestNet4Params)
	if err != nil {
		return fmt.Errorf("Error setting up ltc wallet: \n%s", err)
	}

	adrs, err := btcwallet.AdrDump()
	for i, adr := range adrs {
		fmt.Printf("BTC testnet Wallet addr %d: %s\n", i, base58.CheckEncode(adr[:], 0x6f))
	}

	server.BTCWallet = btcwallet
	server.LTCWallet = ltcwallet
	server.VTCWallet = vtcwallet
	return nil
}

// SetupUserAddress sets up and stores an address for the user, and adds it to the addresses to check for when a block comes in
func (server *OpencxServer) SetupUserAddress(username string) error {
	return nil
}

// TODO: check all cases of new() so you properly delete() and don't have any memory leaks, do this before you start using it

// SetupChainhooks will be used to watch for events on the chain.
func (server *OpencxServer) SetupChainhooks() (btcBlocks chan *wire.MsgBlock, ltcBlocks chan *wire.MsgBlock, vtcBlocks chan *wire.MsgBlock, err error) {
	btcHook := new(uspv.SPVCon)
	ltcHook := new(uspv.SPVCon)
	vtcHook := new(uspv.SPVCon)


	btcHook.Param = &coinparam.TestNet3Params
	ltcHook.Param = &coinparam.LiteCoinTestNet4Params
	vtcHook.Param = &coinparam.VertcoinTestNetParams

	btcRoot := server.createSubRoot(btcHook.Param.Name)
	ltcRoot := server.createSubRoot(ltcHook.Param.Name)
	vtcRoot := server.createSubRoot(vtcHook.Param.Name)

	// Okay now why can I put in "yes" as that parameter or "yup" like that makes no sense as being a remoteNode. "yes" is a remoteNode??
	// maybe isThereAHost should be what its called or something
	btcBlocks = btcHook.RawBlocks()
	btcHook.HardMode = true
	btcHook.Ironman = true
	_, _, err = btcHook.Start(1454277, "1", btcRoot, "", btcHook.Param)
	if err != nil {
		err = fmt.Errorf("Error when starting btc hook: \n%s", err)
		return
	}


	ltcBlocks = ltcHook.RawBlocks()
	ltcHook.HardMode = true
	ltcHook.Ironman = true
	_, _, err = ltcHook.Start(929073, "1", ltcRoot, "", ltcHook.Param)
	if err != nil {
		err = fmt.Errorf("Error when starting ltc hook: \n%s", err)
		return
	}


	vtcBlocks = vtcHook.RawBlocks()
	vtcHook.HardMode = true
	vtcHook.Ironman = true
	_, _, err = vtcHook.Start(186871, "1", vtcRoot, "", vtcHook.Param)
	if err != nil {
		err = fmt.Errorf("Error when starting vtc hook: \n%s", err)
		return
	}

	// server.LTCHook = ltcHook
	// server.BTCHook = btcHook
	// server.VTCHook = vtcHook

	return
}

// HandleBlock handles a block coming in TODO: change coin to not a string if appropriate
func (server *OpencxServer) HandleBlock(block *wire.MsgBlock, coin string) error {
	server.OpencxDB.LogPrintf("Handling block for %s chain\n", coin)
	return nil
}

// createSubRoot creates sub root directories that hold info for each chain
func (server *OpencxServer) createSubRoot(subRoot string) string {
	subRootDir := server.OpencxRoot + subRoot
	if _, err := os.Stat(subRootDir); os.IsNotExist(err) {
		fmt.Printf("Creating root directory at %s\n", subRootDir)
		os.Mkdir(subRootDir, os.ModePerm)
	}
	return subRootDir
}
