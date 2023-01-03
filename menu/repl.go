package menu

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	pb "github.com/libp2p/go-libp2p-core/introspection/pb"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/introspect"
	"github.com/libp2p/go-libp2p/introspect/ws"
	"github.com/libp2p/go-libp2p/p2p/discovery"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/gorilla/websocket"
	"github.com/manifoldco/promptui"
	"github.com/multiformats/go-multiaddr"
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
	wsconn    *websocket.Conn
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

	// Let's build a new libp2p host with introspect enabled. The New constructor uses functional
	// parameters. You don't need to provide any parameters. libp2p comes with
	// sane defaults OOTB, but in order to stay slim, we don't attach a routing
	// implementation by default. Let's do that.
	host, err := libp2p.New(ctx,
		libp2p.Routing(newDHT),
		libp2p.Introspection(
			introspect.NewDefaultIntrospector,
			ws.EndpointWithConfig(&ws.EndpointConfig{ListenAddrs: []string{"127.0.0.1:"}}),
		),
		libp2p.BandwidthReporter(metrics.NewBandwidthCounter()),
	)

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

	host.SetStreamHandler("/taipei/chat/2019", repl.chatProtocolHandler)
	mdns.RegisterNotifee(repl)

	return repl, nil
}

func (r *REPL) Run() {
	type cmd struct {
		name string
		exec func() error
	}

	commands := []cmd{
		{"My info", r.handleMyInfo},
		{"DHT: Bootstrap (public seeds)", func() error { return r.handleDHTBootstrap(dht.DefaultBootstrapPeers...) }},
		{"DHT: Bootstrap (no seeds)", func() error { return r.handleDHTBootstrap() }},
		{"DHT: Print routing table", r.handlePrintRoutingTable},
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

	if err := r.initIntrospectorClient(); err == nil {
		commands = append(commands,
			cmd{"Introspect: Request State", r.handleIntrospectRequest(pb.ClientCommand_STATE)},
			cmd{"Introspect: Request Runtime", r.handleIntrospectRequest(pb.ClientCommand_RUNTIME)},
			cmd{"Introspect: Enable Push Events", r.handleIntrospectEnablePush(pb.ClientCommand_EVENTS)},
			cmd{"Introspect: Enable Push State", r.handleIntrospectEnablePush(pb.ClientCommand_STATE)},
			cmd{"Introspect: Disable Push Events", r.handleIntrospectDisablePush(pb.ClientCommand_EVENTS)},
			cmd{"Introspect: Disable Push State", r.handleIntrospectDisablePush(pb.ClientCommand_STATE)},
			cmd{"Introspect: Pause Push", r.handleIntrospectPausePush},
			cmd{"Introspect: Resume Push", r.handleIntrospectResumePush},
		)
	} else {
		fmt.Println("no introspection actions available: ", err)
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
