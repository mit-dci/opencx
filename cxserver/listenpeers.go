package cxserver

import "github.com/mit-dci/opencx/logging"

// ListenForPeers starts listening for peers using the peer manager
func (server *OpencxServer) ListenForPeers(port uint16) (err error) {

	// again, port shouldn't ever be negative right?
	if err = server.PeerManager.ListenOnPort(int(port)); err != nil {
		return
	}

	lnaddr := server.PeerManager.GetExternalAddress()
	logging.Infof("Listening with ln address: %s\n", lnaddr)

	return
}
