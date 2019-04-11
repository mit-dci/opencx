package cxserver

import (
	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lncore"
	"github.com/mit-dci/lit/lnp2p"
)

// SetupP2PWithDatabase sets up the peer manager with a tracker URL and user specified peer storage.
func (server *OpencxServer) SetupP2PWithDatabase(trackerURL string, peerStorage lncore.LitPeerStorage) (err error) {

	// Create a new event bus TODO: figure out what to actually do here
	ebus := eventbus.NewEventBus()

	if server.PeerManager, err = lnp2p.NewPeerManager(server.peerPrivKey, peerStorage, trackerURL, &ebus, nil); err != nil {
		return
	}

	return
}
