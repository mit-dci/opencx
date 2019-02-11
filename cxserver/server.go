package cxserver

import (
	"fmt"
	"os"
	"sync"

	"github.com/mit-dci/lit/wallit"

	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"

	"github.com/mit-dci/lit/btcutil/hdkeychain"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/lnutil"
	"github.com/mit-dci/lit/wire"
	"github.com/mit-dci/opencx/db/ocxsql"
	"github.com/mit-dci/opencx/logging"
)

// put this here for now, eventually TODO: store stuff as blocks come in and check what height we're at, also deal with reorgs
const exchangeStartingPoint = 1444700
const runningLocally = true

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
	OpencxBTCWallet      *wallit.Wallit
	OpencxLTCWallet      *wallit.Wallit
	OpencxVTCWallet      *wallit.Wallit
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

// SetupBTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupBTCChainhook(errChan chan error) {
	// set log level for this thread
	logging.SetLogLevel(2)

	var btcHost string
	var btcParam *coinparam.Params
	if runningLocally {
		btcHost = "localhost:18444"
		btcParam = &coinparam.RegressionNetParams
		btcParam.StartHeight = 0
	} else {
		btcHost = "1"
		btcParam = &coinparam.TestNet3Params
	}

	btcParam.DiffCalcFunction = dummyDifficulty
	btcRoot := server.createSubRoot(btcParam.Name)

	logging.Infof("Starting BTC Wallet\n")

	btcWallet, coinType, err := wallit.NewWallit(server.OpencxVTCTestPrivKey, btcParam.StartHeight, true, btcHost, btcRoot, "", btcParam)
	if err != nil {
		errChan <- fmt.Errorf("Error when starting btc wallet")
		return
	}

	logging.Infof("BTC Wallet Started, cointype: %d\n", coinType)

	blockChan := btcWallet.Hook.RawBlocks()
	btcHeightChan := btcWallet.LetMeKnowHeight()
	server.OpencxBTCWallet = btcWallet

	go server.HeightHandler(btcHeightChan, blockChan, btcParam)

	errChan <- nil
	return
}

// SetupLTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupLTCChainhook(errChan chan error) {
	// set log level for this thread
	logging.SetLogLevel(2)

	var ltcHost string
	var ltcParam *coinparam.Params
	if runningLocally {
		// TODO: move all this stuff up to be server parameters. Find a way to elegantly manage and add multiple chains while keeping track of parameters
		// and nicely connecting to nodes, while handling unable to connect stuff
		ltcParam = &coinparam.LiteRegNetParams
		ltcHost = "localhost:19444"
		ltcParam.StartHeight = 0
	} else {
		ltcParam = &coinparam.LiteCoinTestNet4Params
		ltcParam.PoWFunction = dummyProofOfWork
		ltcHost = "1"
	}

	// difficulty in non bitcoin testnets has an air of mystery
	ltcParam.DiffCalcFunction = dummyDifficulty

	ltcRoot := server.createSubRoot(ltcParam.Name)

	logging.Infof("Starting LTC Wallet\n")

	ltcWallet, coinType, err := wallit.NewWallit(server.OpencxVTCTestPrivKey, ltcParam.StartHeight, true, ltcHost, ltcRoot, "", ltcParam)
	if err != nil {
		errChan <- fmt.Errorf("Error when starting ltc wallet")
		return
	}

	logging.Infof("LTC Wallet started, coinType: %d\n", coinType)

	blockChan := ltcWallet.Hook.RawBlocks()
	ltcHeightChan := ltcWallet.LetMeKnowHeight()
	server.OpencxLTCWallet = ltcWallet

	go server.HeightHandler(ltcHeightChan, blockChan, ltcParam)

	errChan <- nil
	return
}

// SetupVTCChainhook will be used to watch for events on the chain.
func (server *OpencxServer) SetupVTCChainhook(errChan chan error) {
	// set log level for this thread
	logging.SetLogLevel(2)

	var vtcHost string
	var vtcParam *coinparam.Params
	if runningLocally {
		vtcParam = &coinparam.VertcoinRegTestParams
		vtcHost = "localhost:20444"
		vtcParam.DefaultPort = "20444"
	} else {
		vtcParam = &coinparam.VertcoinTestNetParams
		vtcParam.PoWFunction = dummyProofOfWork
		vtcParam.DNSSeeds = []string{"jlovejoy.mit.edu", "gertjaap.ddns.net", "fr1.vtconline.org", "tvtc.vertcoin.org"}
		vtcHost = "1"
	}

	vtcParam.DiffCalcFunction = dummyDifficulty
	vtcRoot := server.createSubRoot(vtcParam.Name)

	logging.Infof("Starting VTC Wallet\n")

	vtcWallet, coinType, err := wallit.NewWallit(server.OpencxVTCTestPrivKey, vtcParam.StartHeight, true, vtcHost, vtcRoot, "", vtcParam)
	if err != nil {
		errChan <- fmt.Errorf("Error when starting vtc wallet")
		return
	}

	logging.Infof("VTC Wallet started, coinType: %d\n", coinType)

	blockChan := vtcWallet.Hook.RawBlocks()
	vtcHeightChan := vtcWallet.LetMeKnowHeight()
	server.OpencxVTCWallet = vtcWallet

	go server.HeightHandler(vtcHeightChan, blockChan, vtcParam)

	errChan <- nil
	return
}

// TransactionHandler handles incoming transactions
func (server *OpencxServer) TransactionHandler(incomingTxChan chan lnutil.HeightEvent) {
	for {
		logging.Infof("Waiting for incoming transaction...\n")
		txHeight := <-incomingTxChan

		logging.Infof("Got transaction at height %d for cointype %d\n", txHeight.Height, txHeight.CoinType)
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
func (server *OpencxServer) HeightHandler(incomingBlockHeight chan lnutil.HeightEvent, blockChan chan *wire.MsgBlock, coinType *coinparam.Params) {
	for {

		h := <-incomingBlockHeight

		block := <-blockChan
		logging.Debugf("Ingesting %d transactions at height %d\n", len(block.Transactions), h.Height)
		if err := server.ingestTransactionListAndHeight(block.Transactions, uint64(h.Height), coinType); err != nil {
			logging.Infof("something went horribly wrong with %s\n", coinType.Name)
			logging.Errorf("Here's what went horribly wrong: %s\n", err)
		}
	}
}

func dummyDifficulty(headers []*wire.BlockHeader, height int32, p *coinparam.Params) (uint32, error) {
	return headers[len(headers)-1].Bits, nil
}

func dummyProofOfWork(b []byte, height int32) chainhash.Hash {
	lowestPow := make([]byte, 32)
	smolhash, err := chainhash.NewHash(lowestPow)
	if err != nil {
		logging.Errorf("Error setting hash to a bunch of bytes of zeros")
	}
	return *smolhash
}
