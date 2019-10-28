package menu

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/manifoldco/promptui"
)

func (r *REPL) chatProtocolHandler(s network.Stream) {
	fmt.Printf("*** Got a new chat stream from %s! ***\n", s.Conn().RemotePeer())
	r.streams <- s
}

func (r *REPL) handleInitiateChat() error {
	p := promptui.Prompt{Label: "peer id"}
	id, err := p.Run()
	if err != nil {
		return err
	}
	pid, err := peer.IDB58Decode(id)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s, err := r.h.NewStream(ctx, pid, "/taipei/chat/2019")
	if err != nil {
		return err
	}

	return handleChat(s)
}

func (r *REPL) handleAcceptChat() error {
	select {
	case s := <-r.streams:
		return handleChat(s)
	default:
		fmt.Println("no incoming chats")
	}
	return nil
}

func handleChat(s network.Stream) error {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go func() {
		for {
			str, err := rw.ReadString('\n')
			fmt.Printf("<<<< %s\n", str)
			if err != nil {
				fmt.Println("chat closed")
				s.Close()
				return
			}
		}
	}()

	p := promptui.Prompt{Label: "message"}
	for {
		msg, err := p.Run()
		if err != nil {
			return err
		}
		if msg == "." {
			s.Close()
		}
		if _, err := rw.WriteString(msg + "\n"); err != nil {
			return err
		}
		rw.Flush()
	}
}
