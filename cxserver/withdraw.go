package cxserver

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/mit-dci/lit/btcutil/base58"

	"github.com/mit-dci/lit/wire"

	"github.com/mit-dci/lit/btcutil/txscript"

	"github.com/mit-dci/lit/btcutil/chaincfg"

	"github.com/mit-dci/lit/btcutil"

	"github.com/mit-dci/lit/coinparam"
	"github.com/mit-dci/lit/wallit"
	"github.com/mit-dci/opencx/logging"
)

// VTCWalletSend will be used to watch for events on the chain.
func (server *OpencxServer) VTCWalletSend(address string, username string, amount uint64) (string, error) {
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
		return "", fmt.Errorf("Error when deriving child private key for creating withdrawal: \n%s", err)
	}

	vtcRoot := server.createSubRoot(vtcWalletParam.Name)
	logging.Infof("Starting VTC Wallet\n")

	vtcWallet, _, err := wallit.NewWallit(childKey, vtcWalletParam.StartHeight, true, vtcHost, vtcRoot, "", vtcWalletParam)
	if err != nil {
		return "", fmt.Errorf("Error when starting vtc wallet: \n%s", err)
	}
	logging.Infof("VTC Wallet started\n")

	// wrap chainCfg
	chainCfgWrap := &chaincfg.VertcoinTestNetParams
	chainCfgWrap.PubKeyHashAddrID = vtcWalletParam.PubKeyHashAddrID

	childKeyString := childKey.String()
	logging.Infof("Child private key string: %s\n", childKeyString)
	// trying to check that the addresses are the same
	childPubKey, err := childKey.Neuter()
	if err != nil {
		return "", fmt.Errorf("Error when trying to neuter and store key stuffs: \n%s", err)
	}

	logging.Infof("Child pubkey string: %s\n", childPubKey.String())
	pubkey, err := childKey.ECPubKey()
	woah, err := childPubKey.ECPubKey()
	logging.Infof("Other child pubkey (from priv): %x\n", pubkey.SerializeUncompressed())
	logging.Infof("Child pubkey from neutered: %x\n", woah.SerializeUncompressed())

	// childPubKey is the correct pub key to sign with, the non neutered child key is also correct
	addr, err := childPubKey.Address(chainCfgWrap)
	if err != nil {
		return "", err
	}
	logging.Infof("check that the addr is equal to: %s\n", addr.String())

	// for debugging
	addrs, err := vtcWallet.AdrDump()
	for _, addr := range addrs {
		logging.Infof("Address: %s\n", base58.CheckEncode(addr[:], vtcWalletParam.PubKeyHashAddrID))
	}

	// Decoding given address
	vtcAddress, err := btcutil.DecodeAddress(address, chainCfgWrap)
	if err != nil {
		return "", err
	}

	// for paying the other person
	payToUserScript, err := txscript.PayToAddrScript(vtcAddress)
	if err != nil {
		return "", err
	}

	// pick inputs for transaction, idk what the fee shoud be, I think this is the correct byte size but I'm not too sure
	utxoSlice, overshoot, err := vtcWallet.PickUtxos(int64(amount), int64(len(payToUserScript)), 10, false)
	if err != nil {
		return "", err
	}
	logging.Infof("Picked inputs for utxos")

	// for giving back the wallet change
	changeOut, err := vtcWallet.NewChangeOut(overshoot)
	if err != nil {
		return "", err
	}

	// create paytouser txout, we already have change txout from newchangeout
	payToUserTxOut := wire.NewTxOut(int64(amount)-overshoot, payToUserScript)

	// build the transaction
	withdrawTx, err := vtcWallet.BuildAndSign(utxoSlice, []*wire.TxOut{changeOut, payToUserTxOut}, 0)
	if err != nil {
		return "", err
	}
	logging.Infof("Built and signed withdraw transaction")

	// send out the transaction
	if err = vtcWallet.NewOutgoingTx(withdrawTx); err != nil {
		return "", err
	}

	return withdrawTx.TxHash().String(), nil
}
