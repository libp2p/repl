package menu

import "fmt"

func (r *REPL) handleMyInfo() error {
	// 0b. Let's get a sense of what those defaults are. What transports are we
	// listening on? Each transport will have a multiaddr. If you run this
	// multiple times, you will get different port numbers. Note how we listen
	// on all interfaces by default.
	fmt.Println("My addresses:")
	for _, a := range r.h.Addrs() {
		fmt.Printf("\t%s\n", a)
	}

	fmt.Println()
	fmt.Println("My peer ID:")
	fmt.Printf("\t%s\n", r.h.ID())

	fmt.Println()
	fmt.Println("My identified multiaddrs:")
	for _, a := range r.h.Addrs() {
		fmt.Printf("\t%s/p2p/%s\n", a, r.h.ID())
	}

	// What protocols are added by default?
	fmt.Println()
	fmt.Println("Protocols:")
	for _, p := range r.h.Mux().Protocols() {
		fmt.Printf("\t%s\n", p)
	}

	// What peers do we have in our peerstore? (hint: we've connected to nobody so far).
	fmt.Println()
	fmt.Println("Peers in peerstore:")
	for _, p := range r.h.Peerstore().PeersWithAddrs() {
		fmt.Printf("\t%s\n", p)
	}

	return nil
}
