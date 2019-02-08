package cxserver

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wallit"
	"github.com/mit-dci/opencx/logging"
)

// WithdrawToAddress is called by the RPC handler and publishes a transaction, returning the txid for the withdrawal
func (server *OpencxServer) WithdrawToAddress(address string, account string, chain string, amount uint64) (txid string, err error) {

	return "", nil
}

// VTCWalletSend will be used to watch for events on the chain.
func (server *OpencxServer) VTCWalletSend(address string, username string, amount uint64) error {
	// set log level for this thread
	logging.SetLogLevel(2)

	var vtcWalletParam *coinparam.Params
	var vtcHost string
	if runningLocally {
		vtcWalletParam = &coinparam.VertcoinRegTestParams
		vtcHost = "localhost:20444"
		vtcWalletParam.DefaultPort = "20444"
	} else {
		vtcWalletParam = &coinparam.VertcoinTestNetParams
		vtcWalletParam.PoWFunction = dummyProofOfWork
		vtcWalletParam.DNSSeeds = []string{"jlovejoy.mit.edu", "gertjaap.ddns.net", "fr1.vtconline.org", "tvtc.vertcoin.org"}
		vtcHost = "1"
	}

	// creating derived key from root
	sha := sha256.New()
	sha.Write([]byte(username))

	// TODO: Make this better lol, one of the most annoying things (and now that I think about it, exchanges must have this down to a T) is key assignment and management
	// We mod by 0x80000000 to make sure it's not hardened
	data := binary.BigEndian.Uint32(sha.Sum(nil)[:]) % 0x80000000

	childKey, err := server.OpencxVTCTestPrivKey.Child(data)
	if err != nil {
		return fmt.Errorf("Error when deriving child private key for creating withdrawal: \n%s", err)
	}

	vtcRoot := server.createSubRoot(vtcWalletParam.Name)
	logging.Infof("Starting VTC Wallet\n")

	vtcWallet, _, err := wallit.NewWallit(childKey, vtcWalletParam.StartHeight, true, vtcHost, vtcRoot, "", vtcWalletParam)
	if err != nil {
		return fmt.Errorf("Error when starting vtc wallet: \n%s", err)
	}
	logging.Infof("VTC Wallet started\n")

	// idk the output byte size so something like 512
	_, _, err = vtcWallet.PickUtxos(int64(amount), int64(512), 10, false)
	if err != nil {
		return err
	}

	return nil
}
