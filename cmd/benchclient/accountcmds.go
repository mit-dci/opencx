package benchclient

import (
	"fmt"

	"github.com/mit-dci/opencx/cxrpc"
)

// Register registers for an account
func (cl *BenchClient) Register(signature []byte) (registerReply *cxrpc.RegisterReply, err error) {

	if cl.PrivKey == nil {
		err = fmt.Errorf("Private key nonexistent, set or specify private key so the client can sign commands")
		return
	}

	// sign it

	// now send it back to prove your knowledge of discrete logarithm of your public key, AKA Prove you know your privkey by signing this message
	registerReply = new(cxrpc.RegisterReply)
	registerArgs := &cxrpc.RegisterArgs{
		Signature: signature,
	}

	if err = cl.Call("OpencxRPC.Register", registerArgs, registerReply); err != nil {
		return
	}

	return
}

// GetRegistrationString gets the registration string that needs to be signed in order to be registered on the exchange
func (cl *BenchClient) GetRegistrationString() (getRegistrationStringReply *cxrpc.GetRegistrationStringReply, err error) {

	getRegistrationStringReply = new(cxrpc.GetRegistrationStringReply)
	getRegistrationStringArgs := &cxrpc.GetRegistrationStringArgs{}

	if err = cl.Call("OpencxRPC.GetRegistrationString", getRegistrationStringArgs, getRegistrationStringReply); err != nil {
		return
	}

	return
}
