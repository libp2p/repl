package menu

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (r *REPL) HandlePeerFound(pi peer.AddrInfo) {
	r.m.Lock()
	r.mdnsPeers[pi.ID] = pi
	r.m.Unlock()

	r.h.Connect(r.ctx, pi)
}

func (r *REPL) handleListmDNSPeers() error {
	r.m.RLock()
	defer r.m.RUnlock()

	spew.Dump(r.mdnsPeers)
	return nil
}
