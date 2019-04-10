package cxserver

import (
	"fmt"

	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lncore"
	"github.com/mit-dci/lit/lnp2p"
)

// SetupPeerToPeer is super unfinished, trying to think about architecture because lit peermgr might not be the
// best thing to use. Also figure out why we're even using trackerURL.
func (server *OpencxServer) SetupPeerToPeer(trackerURL string) (err error) {
	// TODO
	var opencxPeerStorage lncore.LitPeerStorage
	var correctType bool
	if opencxPeerStorage, correctType = server.OpencxDB.(lncore.LitPeerStorage); !correctType {
		err = fmt.Errorf("The OpenCX Database does not have the required methods to serve as a database for p2p networking. Use another method to set up p2p networking or fix your database methods")
		return
	}

	// Create a new event bus TODO: figure out what to actually do here
	ebus := eventbus.NewEventBus()

	if server.PeerManager, err = lnp2p.NewPeerManager(server.peerPrivKey, opencxPeerStorage, trackerURL, &ebus, nil); err != nil {
		return
	}

	return
}

// SetupP2PWithDatabase sets up the peer manager with a tracker URL and user specified peer storage.
func (server *OpencxServer) SetupP2PWithDatabase(trackerURL string, peerStorage lncore.LitPeerStorage) (err error) {

	// Create a new event bus TODO: figure out what to actually do here
	ebus := eventbus.NewEventBus()

	if server.PeerManager, err = lnp2p.NewPeerManager(server.peerPrivKey, peerStorage, trackerURL, &ebus, nil); err != nil {
		return
	}

	return
}
