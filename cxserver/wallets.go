package cxserver

import (
	"fmt"
	"os"

	"github.com/mit-dci/lit/btcutil"
	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wallit"
	util "github.com/mit-dci/opencx/chainutils"

	"github.com/mit-dci/opencx/logging"
)

// SetupWallet sets up a wallet for a specific coin, based on params.
func (server *OpencxServer) SetupWallet(errChan chan error, subDirName string, param *coinparam.Params, resync bool, hostString string) {
	var err error
	var coinType int
	defer func() {
		if err != nil {
			err = fmt.Errorf("Error when starting wallet: \n%s", err)
		}
		errChan <- err
	}()

	logging.Infof("Starting %s wallet\n", param.Name)

	key, found := server.PrivKeyMap[param]
	if !found {
		err = fmt.Errorf("Could not find key for wallet. Aborting wallet setup")
		return
	}

	// create wallit root directory
	if _, err = os.Stat(server.OpencxRoot + subDirName); os.IsNotExist(err) {
		if err = os.Mkdir(server.OpencxRoot+subDirName, 0700); err != nil {
			logging.Errorf("Error while creating a directory: \n%s", err)
		}
	}

	var wallet *wallit.Wallit
	if wallet, coinType, err = wallit.NewWallit(key, param.StartHeight, resync, hostString, server.OpencxRoot+subDirName, "", param); err != nil {
		return
	}

	server.walletMtx.Lock()
	server.WalletMap[param] = wallet
	server.walletMtx.Unlock()

	logging.Infof("%s wallet Started, cointype: %d\n", param.Name, coinType)
	// figure out whether or not to do this if merged

	server.StartChainhookHandlers(wallet)

	return
}

// SetupAllWallets sets up all wallets with parameters as specified in the hostParamList
func (server *OpencxServer) SetupAllWallets(hostParamList util.HostParamList, subDirName string, resync bool) (err error) {
	hpLen := len(hostParamList)
	errChan := make(chan error, hpLen)
	for _, hostParam := range hostParamList {
		go server.SetupWallet(errChan, subDirName, hostParam.Param, resync, hostParam.Host)
	}

	for i := 0; i < hpLen; i++ {
		if err = <-errChan; err != nil {
			return
		}
	}
	return
}

// LinkAllWallets will link the exchanges' wallets with the lit node running. Defaults to false for running tower.
func (server *OpencxServer) LinkAllWallets() (err error) {

	// Not sure whether or not this should just assume that everything in the map is what you want, but I'm going to
	// assume that if there's a coin / param in the CoinList that isn't in the wallet map, then the wallets haven't
	// been started or something is wrong. This is definitely a synchronous thing to be doing, you need to start
	// the wallets for all your coins before you try to link them all. If you don't want to link them all, use
	// LinkManyWallets.

	// Check for all settlement layers
	for param, _ := range server.SettlementEngines {
		wallet, found := server.WalletMap[param]
		if !found {
			err = fmt.Errorf("Wallet in Coin List not being tracked by exchange in map, start it please")
		}

		// Idk if I should run a tower with these, probably. It's an exchange
		if err = server.LinkOneWallet(wallet, false); err != nil {
			return
		}
	}

	logging.Infof("Successfully linked all wallets!")
	return
}

// LinkManyWallets takes in a bunch of wallets and links them. We set the tower for all wallets consistently
// coinType as specified in the parameters.
func (server *OpencxServer) LinkManyWallets(wallets []*wallit.Wallit, tower bool) (err error) {
	for _, wallet := range wallets {
		if err = server.LinkOneWallet(wallet, tower); err != nil {
			return
		}
	}

	return
}

// LinkOneWallet is a modified version of linkwallet in lit that doesn't make the wallet but links it with an already running one. Your responsibility to pass the correct cointype and tower.
func (server *OpencxServer) LinkOneWallet(wallet *wallit.Wallit, tower bool) (err error) {
	// we don't need param passed as a parameter to this function, the wallet already has it so we have to substitute a bunch of stuff
	WallitIdx := wallet.Param.HDCoinType

	// see if we've already attached a wallet for this coin type
	if server.ExchangeNode.SubWallet[WallitIdx] != nil {
		err = fmt.Errorf("coin type %d already linked", WallitIdx)
		return
	}

	// see if there are other wallets already linked
	if len(server.ExchangeNode.SubWallet) != 0 {
		// there are; assert multiwallet (may already be asserted)
		server.ExchangeNode.MultiWallet = true
	}

	// Have to do this because we deleted the lines actually creating the wallet, we already have it created.
	server.ExchangeNode.SubWallet[WallitIdx] = wallet

	// if there aren't, Multiwallet will still be false; set new wallit to
	// be the first & default

	if server.ExchangeNode.ConnectedCoinTypes == nil {
		server.ExchangeNode.ConnectedCoinTypes = make(map[uint32]bool)
	}

	// why is this needed in 2 places, can't this be the only time this is run?
	server.ExchangeNode.ConnectedCoinTypes[WallitIdx] = true

	// re-register channel addresses
	qChans, err := server.ExchangeNode.GetAllQchans()
	if err != nil {
		return err
	}

	for _, qChan := range qChans {
		var pkh [20]byte
		pkhSlice := btcutil.Hash160(qChan.MyRefundPub[:])
		copy(pkh[:], pkhSlice)
		server.ExchangeNode.SubWallet[WallitIdx].ExportHook().RegisterAddress(pkh)

		logging.Infof("Registering outpoint %v", qChan.PorTxo.Op)

		server.ExchangeNode.SubWallet[WallitIdx].WatchThis(qChan.PorTxo.Op)
	}

	go server.ExchangeNode.OPEventHandler(server.ExchangeNode.SubWallet[WallitIdx].LetMeKnow())
	go server.ExchangeNode.HeightEventHandler(server.HeightEventChanMap[int(WallitIdx)])

	// If this is the first coin we're linking then set that one to default.
	if !server.ExchangeNode.MultiWallet {
		server.ExchangeNode.DefaultCoin = wallet.Param.HDCoinType
	}

	// if this node is running a watchtower, link the watchtower to the
	// new wallet block events

	if tower {
		err = server.ExchangeNode.Tower.HookLink(
			server.ExchangeNode.LitFolder, wallet.Param, server.ExchangeNode.SubWallet[WallitIdx].ExportHook())
		if err != nil {
			return err
		}
	}

	return nil
}
