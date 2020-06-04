package menu

import (
	"fmt"
	"sync/atomic"

	"github.com/libp2p/go-libp2p-core/host"
	pb "github.com/libp2p/go-libp2p-core/introspection/pb"

	"github.com/gorilla/websocket"
	"github.com/logrusorgru/aurora"
)

var counter uint64

func (r *REPL) initIntrospectorClient() error {
	ih, ok := r.h.(host.IntrospectableHost)
	if !ok {
		return fmt.Errorf("host is not introspectable")
	}

	// create connection to WS introspect server
	addrs := ih.IntrospectionEndpoint().ListenAddrs()
	url := fmt.Sprintf("ws://%s/introspect", addrs[0])

	fmt.Printf("introspection server running at: %s\n", url)

	wsconn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}

	r.wsconn = wsconn

	// prepare hello
	hello := &pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_HELLO}
	bytes, err := hello.Marshal()
	if err != nil {
		return err
	}

	// send hello
	if err := wsconn.WriteMessage(websocket.BinaryMessage, bytes); err != nil {
		return err
	}

	// reader loop.
	go func() {
		for {
			_, bytes, err := wsconn.ReadMessage()
			if err != nil {
				fmt.Println("failed to read introspection message; yielding reader loop: ", err)
				return
			}

			msg := new(pb.ServerMessage)
			bytes = bytes[12:] // discard header
			if err := msg.Unmarshal(bytes); err != nil {
				fmt.Println("failed to unmarshal introspection message: ", err)
				continue
			}

			fmt.Println(aurora.Blue("👀 " + msg.String()))
		}
	}()

	return nil
}

func (r *REPL) handleIntrospectRequest(source pb.ClientCommand_Source) func() error {
	return func() error {
		cmd := &pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_REQUEST, Source: source}
		return r.sendCommand(cmd)
	}
}

func (r *REPL) handleIntrospectEnablePush(source pb.ClientCommand_Source) func() error {
	return func() error {
		cmd := &pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_PUSH_ENABLE, Source: source}
		return r.sendCommand(cmd)
	}
}

func (r *REPL) handleIntrospectDisablePush(source pb.ClientCommand_Source) func() error {
	return func() error {
		cmd := &pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_PUSH_DISABLE, Source: source}
		return r.sendCommand(cmd)
	}
}

func (r *REPL) handleIntrospectPausePush() error {
	return r.sendCommand(&pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_PUSH_PAUSE})
}

func (r *REPL) handleIntrospectResumePush() error {
	return r.sendCommand(&pb.ClientCommand{Id: atomic.AddUint64(&counter, 1), Command: pb.ClientCommand_PUSH_RESUME})
}

func (r *REPL) sendCommand(cmd *pb.ClientCommand) error {
	if bytes, err := cmd.Marshal(); err == nil {
		return r.wsconn.WriteMessage(websocket.BinaryMessage, bytes)
	} else {
		return err
	}
}
