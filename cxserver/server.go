package cxserver

import (
	"fmt"
	"os"
	"sync"

	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/uspv"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/util"
)

// put this here for now, eventually TODO: store stuff as blocks come in and check what height we're at, also deal with reorgs
const exchangeStartingPoint = 1444700

// OpencxServer is how rpc can query the database and whatnot
type OpencxServer struct {
	OpencxDB   *ocxsql.DB
	OpencxRoot string
	OpencxPort int
	// Hehe it's the vault, pls don't steal
	OpencxBTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxVTCTestPrivKey *hdkeychain.ExtendedKey
	OpencxLTCTestPrivKey *hdkeychain.ExtendedKey
	HeightEventChan      chan lnutil.HeightEvent
	ingestMutex          sync.Mutex
	// TODO: Or implement client required signatures and pubkeys instead of usernames
}

// TODO now that I know how to use this hdkeychain stuff, let's figure out how to create addresses to store

// SetupServerKeys just loads a private key from a file wallet
func (server *OpencxServer) SetupServerKeys(keypath string) error {
	privkey, err := lnutil.ReadKeyFile(keypath)
	if err != nil {
		return fmt.Errorf("Error reading key from file: \n%s", err)
	}

	server.OpencxDB.Keychain = new(util.Keychain)

	rootBTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.TestNet3Params)
	if err != nil {
		return fmt.Errorf("Error creating master BTC Test key from private key: \n%s", err)
	}

	server.OpencxBTCTestPrivKey = rootBTCKey
	server.OpencxDB.Keychain.BTCPubKey, err = rootBTCKey.Neuter()
	if err != nil {
		return fmt.Errorf("Error neutering btc privkey while setting up keys: \n%s", err)
	}

	rootVTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.VertcoinRegTestParams)
	if err != nil {
		return fmt.Errorf("Error creating master VTC Test key from private key: \n%s", err)
	}

	server.OpencxVTCTestPrivKey = rootVTCKey
	server.OpencxDB.Keychain.VTCPubKey, err = rootVTCKey.Neuter()
	if err != nil {
		return fmt.Errorf("Error neutering btc privkey while setting up keys: \n%s", err)
	}

	rootLTCKey, err := hdkeychain.NewMaster(privkey[:], &coinparam.LiteCoinTestNet4Params)
	if err != nil {
		return fmt.Errorf("Error creating master LTC Test key from private key: \n%s", err)
	}

	server.OpencxLTCTestPrivKey = rootLTCKey
	server.OpencxDB.Keychain.LTCPubKey, err = rootLTCKey.Neuter()
	if err != nil {
		return fmt.Errorf("Error neutering btc privkey while setting up keys: \n%s", err)
	}

	return nil
}

// SetupBTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupBTCChainhook() error {
	btcHook := new(uspv.SPVCon)
	logging.SetLogLevel(3)

	btcHook.Param = &coinparam.RegressionNetParams
	if btcHook.Param == &coinparam.RegressionNetParams {
		btcHook.Param.DiffCalcFunction = dummyDifficulty
		btcHook.Param.StartHeight = 0
	}

	btcRoot := server.createSubRoot(btcHook.Param.Name)

	// Okay now why can I put in "yes" as that parameter or "yup" like that makes no sense as being a remoteNode. "yes" is a remoteNode??
	// maybe isThereAHost should be what its called or something
	logging.Debugf("Starting BTC Chainhook\n")
	blockChan := btcHook.RawBlocks()

	// 1454000 is recent enough to not take too long. Also, the addresses weren't made before then so unless we want to
	// credit people from the past idk what the point is

	// I'm also trying to figure out how well full nodes work with chainhooks, don't wanna be a leech but rn it's not working too well
	txHeightChan, btcheightChan, err := btcHook.Start(btcHook.Param.StartHeight, "localhost:18444", btcRoot, "", btcHook.Param)
	if err != nil {
		return fmt.Errorf("Error when starting btc hook: \n%s", err)
	}
	btcHook.Ironman = true
	logging.Infof("BTC Chainhook started\n")

	go server.TransactionHandler(txHeightChan)
	go server.HeightHandler(btcheightChan, blockChan, btcHook.Param)

	return nil
}

// SetupLTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupLTCChainhook() error {
	// logging.SetLogLevel(3)
	ltcHook := new(uspv.SPVCon)

	ltcHook.Param = &coinparam.LiteRegNetParams
	// difficulty in non bitcoin testnets has an air of mystery
	if ltcHook.Param.TestCoin {
		ltcHook.Param.DiffCalcFunction = dummyDifficulty
		ltcHook.Param.StartHeight = 0
	}
	ltcRoot := server.createSubRoot(ltcHook.Param.Name)

	logging.Debugf("Starting LTC Chainhook\n")
	blockChan := ltcHook.RawBlocks()

	txHeightChan, ltcheightChan, err := ltcHook.Start(ltcHook.Param.StartHeight, "localhost:19444", ltcRoot, "", ltcHook.Param)
	if err != nil {
		return fmt.Errorf("Error when starting ltc hook: \n%s", err)
	}
	ltcHook.Ironman = true
	logging.Infof("LTC Chainhook started\n")

	go server.TransactionHandler(txHeightChan)
	go server.HeightHandler(ltcheightChan, blockChan, ltcHook.Param)

	return nil
}

// SetupVTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupVTCChainhook() error {
	vtcHook := new(uspv.SPVCon)
	logging.SetLogLevel(2)

	vtcHook.Param = &coinparam.VertcoinRegTestParams
	if vtcHook.Param.TestCoin {
		vtcHook.Param.DNSSeeds = []string{"jlovejoy.mit.edu", "gertjaap.ddns.net", "fr1.vtconline.org", "tvtc.vertcoin.org"}
	}
	vtcHook.Param.DefaultPort = "20444"
	vtcHook.Param.GenerateSupported = true
	vtcRoot := server.createSubRoot(vtcHook.Param.Name)

	logging.Infof("Starting VTC Chainhook\n")
	blockChan := vtcHook.RawBlocks()

	// vertcoin regtest uses the same port as bitcoin regtest... >:(
	txHeightChan, vtcheightChan, err := vtcHook.Start(vtcHook.Param.StartHeight, "localhost:20444", vtcRoot, "", vtcHook.Param)
	if err != nil {
		return fmt.Errorf("Error when starting vtc hook: \n%s", err)
	}
	vtcHook.Ironman = true
	logging.Debugf("VTC Chainhook started\n")

	go server.TransactionHandler(txHeightChan)
	go server.HeightHandler(vtcheightChan, blockChan, vtcHook.Param)

	return nil
}

// TransactionHandler handles incoming transactions
func (server *OpencxServer) TransactionHandler(incomingTxChan chan lnutil.TxAndHeight) {
	for {
		logging.Infof("Waiting for incoming transaction...\n")
		txHeight := <-incomingTxChan

		logging.Infof("Got transaction at height: %d, txid: %s, outputs: %d\n", txHeight.Height, txHeight.Tx.TxHash().String(), len(txHeight.Tx.TxOut))
	}
}

// createSubRoot creates sub root directories that hold info for each chain
func (server *OpencxServer) createSubRoot(subRoot string) string {
	subRootDir := server.OpencxRoot + subRoot
	if _, err := os.Stat(subRootDir); os.IsNotExist(err) {
		logging.Infof("Creating root directory at %s\n", subRootDir)
		os.Mkdir(subRootDir, os.ModePerm)
	}
	return subRootDir
}

// HeightHandler is a handler for when there is a height and block event. We need both channels to work and be synchronized, which I'm assuming is the case in the lit repos. Will need to double check.
func (server *OpencxServer) HeightHandler(incomingBlockHeight chan int32, blockChan chan *wire.MsgBlock, coinType *coinparam.Params) {
	for {

		h := <-incomingBlockHeight

		block := <-blockChan
		logging.Debugf("Ingesting %d transactions at height %d\n", len(block.Transactions), h)
		// Wow we all have little hope that the bitcoin blockheight will grow to be a 64 bit integer... I want to see the day & hope we have
		// hard drives big enough to hold the entire chain (or just the entire utreexo)
		if err := server.ingestTransactionListAndHeight(block.Transactions, uint64(h), coinType); err != nil {
			logging.Infof("something went horribly wrong with %s", coinType.Name)
		}

	}
}

func dummyDifficulty(headers []*wire.BlockHeader, height int32, p *coinparam.Params) (uint32, error) {
	return headers[0].Bits, nil
}
