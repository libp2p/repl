package menu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	disc "github.com/libp2p/go-libp2p-discovery"
	"github.com/multiformats/go-multiaddr"
)

func (r *REPL) handleBootstrapMode() error {
	fmt.Println("this node will now serve as a DHT bootstrap node, addrs:")
	fmt.Println("peer ID:", r.h.ID())
	fmt.Println("addrs:", r.h.Addrs())
	time.Sleep(24 * time.Hour)
	return nil
}

func (r *REPL) handleDHTBootstrap(seeds ...multiaddr.Multiaddr) error {
	fmt.Println("Will bootstrap for 30 seconds...")

	ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(seeds))

	for _, ma := range seeds {
		ai, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			return err
		}

		go func(ai peer.AddrInfo) {
			defer wg.Done()

			fmt.Printf("Connecting to peer: %s\n", ai)
			if err := r.h.Connect(ctx, ai); err != nil {
				fmt.Printf("Failed while connecting to peer: %s; %s\n", ai, err)
			} else {
				fmt.Printf("Succeeded while connecting to peer: %s\n", ai)
			}
		}(*ai)
	}

	wg.Wait()

	select {
	case <-r.dht.RefreshRoutingTable():
	case <-ctx.Done():
	}

	fmt.Println("bootstrap OK! Routing table:")
	r.dht.RoutingTable().Print()

	return nil
}

func (r *REPL) handleAnnounceService() error {
	rd := disc.NewRoutingDiscovery(r.dht)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := rd.Advertise(ctx, "taipei2019", disc.TTL(10*time.Minute))
	return err
}

func (r *REPL) handleFindProviders() error {
	rd := disc.NewRoutingDiscovery(r.dht)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	peers, err := rd.FindPeers(ctx, "taipei2019")
	if err != nil {
		return err
	}

	for p := range peers {
		fmt.Println("found peer", p)
		r.h.Peerstore().AddAddrs(p.ID, p.Addrs, 24*time.Hour)
	}
	return err
}
