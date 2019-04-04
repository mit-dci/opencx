package benchclient

import (
	"fmt"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/cxrpc"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// GetBalance calls the getbalance rpc command
func (cl *BenchClient) GetBalance(asset string) (getBalanceReply *cxrpc.GetBalanceReply, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	getBalanceReply = new(cxrpc.GetBalanceReply)
	getBalanceArgs := &cxrpc.GetBalanceArgs{
		Asset: asset,
	}

	// create e = hash(m)
	sha3 := sha3.New256()
	sha3.Write([]byte(asset))
	e := sha3.Sum(nil)

	// Sign
	compactSig, err := koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false)

	// set signature in args
	getBalanceArgs.Signature = compactSig

	if err = cl.Call("OpencxRPC.GetBalance", getBalanceArgs, getBalanceReply); err != nil {
		return
	}

	return
}

// GetDepositAddress calls the getdepositaddress rpc command
func (cl *BenchClient) GetDepositAddress(asset string) (getDepositAddressReply *cxrpc.GetDepositAddressReply, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	getDepositAddressReply = new(cxrpc.GetDepositAddressReply)
	getDepositAddressArgs := &cxrpc.GetDepositAddressArgs{
		Asset: asset,
	}

	// create e = hash(m)
	sha3 := sha3.New256()
	sha3.Write([]byte(asset))
	e := sha3.Sum(nil)

	// Sign order
	compactSig, err := koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false)

	// set signature in args
	getDepositAddressArgs.Signature = compactSig

	if err = cl.Call("OpencxRPC.GetDepositAddress", getDepositAddressArgs, getDepositAddressReply); err != nil {
		return
	}

	return
}

// GetAllBalances get the balance for every token
func (cl *BenchClient) GetAllBalances() (balances map[string]uint64, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	balances = make(map[string]uint64)
	var reply *cxrpc.GetBalanceReply
	if reply, err = cl.GetBalance("regtest"); err != nil {
		return
	}

	balances["regtest"] = reply.Amount

	if reply, err = cl.GetBalance("vtcreg"); err != nil {
		return
	}
	balances["vtcreg"] = reply.Amount

	if reply, err = cl.GetBalance("litereg"); err != nil {
		return
	}
	balances["litereg"] = reply.Amount

	return
}

// Withdraw calls the withdraw rpc command
func (cl *BenchClient) Withdraw(amount uint64, asset match.Asset, address string) (withdrawReply *cxrpc.WithdrawReply, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	withdrawReply = new(cxrpc.WithdrawReply)
	withdrawArgs := &cxrpc.WithdrawArgs{
		Withdrawal: &match.Withdrawal{
			Amount:    amount,
			Asset:     asset,
			Address:   address,
			Lightning: false,
		},
	}

	// create e = hash(m)
	sha3 := sha3.New256()
	sha3.Write(withdrawArgs.Withdrawal.Serialize())
	e := sha3.Sum(nil)

	// Sign order
	compactSig, err := koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false)

	// set signature in args
	withdrawArgs.Signature = compactSig

	if err = cl.Call("OpencxRPC.Withdraw", withdrawArgs, withdrawReply); err != nil {
		return
	}

	if withdrawReply.Txid == "" {
		err = fmt.Errorf("Error: Unsupported Asset")
		return
	}

	return
}

// WithdrawLightning calls the withdraw rpc command, but with the lightning boolean set to true
func (cl *BenchClient) WithdrawLightning(amount uint64, asset match.Asset) (withdrawReply *cxrpc.WithdrawReply, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	withdrawReply = new(cxrpc.WithdrawReply)
	withdrawArgs := &cxrpc.WithdrawArgs{
		Withdrawal: &match.Withdrawal{
			Amount:    amount,
			Asset:     asset,
			Lightning: true,
		},
	}

	// create e = hash(m)
	sha3 := sha3.New256()
	sha3.Write(withdrawArgs.Withdrawal.Serialize())
	e := sha3.Sum(nil)

	// Sign order
	compactSig, err := koblitz.SignCompact(koblitz.S256(), cl.PrivKey, e, false)

	// set signature in args
	withdrawArgs.Signature = compactSig

	if err = cl.Call("OpencxRPC.Withdraw", withdrawArgs, withdrawReply); err != nil {
		return
	}

	if withdrawReply.Txid == "" {
		err = fmt.Errorf("Error: Unsupported Asset")
		return
	}

	return
}
