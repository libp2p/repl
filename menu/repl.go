package menu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	protocol "github.com/libp2p/go-libp2p-protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	routing "github.com/libp2p/go-libp2p-routing"
	"github.com/libp2p/go-libp2p/p2p/discovery"

	"github.com/multiformats/go-multiaddr"

	"github.com/manifoldco/promptui"
)

type REPL struct {
	m sync.RWMutex

	ctx      context.Context
	cancelFn func()
	h        host.Host
	dht      *dht.IpfsDHT
	pubsub   *pubsub.PubSub

	mdnsPeers map[peer.ID]peer.AddrInfo
	messages  map[string][]*pubsub.Message
	streams   chan network.Stream
}

func NewREPL() (*REPL, error) {
	// Contexts are an ugly way of controlling component lifecycles.
	// Service-based host refactor upcoming.
	ctx, cancel := context.WithCancel(context.Background())

	var kaddht *dht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		kaddht, err = dht.New(ctx, h)
		return kaddht, err
	}

	// Let's build a new libp2p host. The New constructor uses functional
	// parameters. You don't need to provide any parameters. libp2p comes with
	// sane defaults OOTB, but in order to stay slim, we don't attach a routing
	// implementation by default. Let's do that.
	host, err := libp2p.New(ctx, libp2p.Routing(newDHT))
	if err != nil {
		cancel()
		return nil, err
	}

	mdns, err := discovery.NewMdnsService(ctx, host, time.Second*5, "")
	if err != nil {
		cancel()
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		cancel()
		return nil, err
	}

	repl := &REPL{
		ctx:       ctx,
		cancelFn:  cancel,
		h:         host,
		dht:       kaddht,
		mdnsPeers: make(map[peer.ID]peer.AddrInfo),
		messages:  make(map[string][]*pubsub.Message),
		streams:   make(chan network.Stream, 128),
		pubsub:    ps,
	}

	host.SetStreamHandler(protocol.ID("/taipei/chat/2019"), repl.chatProtocolHandler)
	mdns.RegisterNotifee(repl)

	return repl, nil
}

func (r *REPL) Run() {
	commands := []struct {
		name string
		exec func() error
	}{
		{"My info", r.handleMyInfo},
		{"DHT: Bootstrap (public seeds)", func() error { return r.handleDHTBootstrap(dht.DefaultBootstrapPeers...) }},
		{"DHT: Bootstrap (no seeds)", func() error { return r.handleDHTBootstrap() }},
		{"DHT: Announce service", r.handleAnnounceService},
		{"DHT: Find service providers", r.handleFindProviders},
		{"Network: Connect to a peer", r.handleConnect},
		{"Network: List connections", r.handleListConnectedPeers},
		{"mDNS: List local peers", r.handleListmDNSPeers},
		{"Pubsub: Subscribe to topic", r.handleSubscribeToTopic},
		{"Pubsub: Publish a message", r.handlePublishToTopic},
		{"Pubsub: Print inbound messages", r.handlePrintInboundMessages},
		{"Protocol: Initiate chat with peer", r.handleInitiateChat},
		{"Protocol: Accept incoming chat", r.handleAcceptChat},
		{"Identify peer protocols", r.handleIdentifyPeer},
		{"Switch to bootstrap mode", r.handleBootstrapMode},
	}

	var str []string
	for _, c := range commands {
		str = append(str, c.name)
	}

	for {
		sel := promptui.Select{
			Label: "What do you want to do?",
			Items: str,
			Size:  1000,
		}

		fmt.Println()
		i, _, err := sel.Run()
		if err != nil {
			fmt.Println("shutting down")
			return
		}

		if err := commands[i].exec(); err != nil {
			fmt.Printf("command failed: %s\n", err)
		}
	}
}

func (r *REPL) Close() error {
	defer r.cancelFn()

	if err := r.h.Close(); err != nil {
		return err
	}

	return nil
}

func validateMultiaddr(s string) error {
	ma, err := multiaddr.NewMultiaddr(s)
	if err != nil {
		return err
	}
	_, err = peer.AddrInfoFromP2pAddr(ma)
	return err
}
