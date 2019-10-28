package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/manifoldco/promptui"
	"github.com/multiformats/go-multiaddr"
)

func (r *REPL) handleConnect() error {
	p := promptui.Prompt{
		Label:    "multiaddr",
		Validate: validateMultiaddr,
	}
	addr, err := p.Run()
	if err != nil {
		return err
	}
	ma, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return err
	}
	ai, err := peer.AddrInfoFromP2pAddr(ma)
	ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
	defer cancel()
	return r.h.Connect(ctx, *ai)
}

func (t *REPL) handleListConnectedPeers() error {
	for _, c := range t.h.Network().Conns() {
		fmt.Println("connected to", c.RemotePeer(), "on", c.RemoteMultiaddr())
	}

	return nil
}
