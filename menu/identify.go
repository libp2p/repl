package menu

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/manifoldco/promptui"
)

func (r *REPL) handleIdentifyPeer() error {
	p := promptui.Prompt{Label: "peer id"}
	id, err := p.Run()
	if err != nil {
		return err
	}

	pid, err := peer.IDB58Decode(id)
	if err != nil {
		return err
	}

	if r.h.Network().Connectedness(pid) != network.Connected {
		ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
		defer cancel()

		doneCh := make(chan struct{})

		// Set up a notifee to track identify events. This should be as simple as
		// listening in the event bus, but identify is not emitting events on
		// initial identification yet.
		fo := func(n network.Network, s network.Stream) {
			if s.Conn().RemotePeer() != pid || s.Protocol() != identify.ID {
				return
			}
			fmt.Println("opened identify stream to peer")
		}
		fc := func(n network.Network, s network.Stream) {
			if s.Conn().RemotePeer() != pid || s.Protocol() != identify.ID {
				return
			}
			close(doneCh)
		}
		notifee := &network.NotifyBundle{
			OpenedStreamF: fo,
			ClosedStreamF: fc,
		}

		r.h.Network().Notify(notifee)
		defer r.h.Network().StopNotify(notifee)

		fmt.Println(ctx)

		err = r.h.Connect(ctx, peer.AddrInfo{ID: pid})
		if err != nil {
			return err
		}

		select {
		case <-doneCh:
		case <-time.After(10 * time.Second):
			return fmt.Errorf("unable to identify peer %s in a reasonable time", pid)
		}
	}

	protos, err := r.h.Peerstore().GetProtocols(pid)
	if err != nil {
		return err
	}
	fmt.Println(protos)
	return nil

}
