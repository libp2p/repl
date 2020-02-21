package menu

import (
	"errors"
	"fmt"

	introspection_pb "github.com/libp2p/go-libp2p-core/introspection/pb"
	"github.com/libp2p/go-libp2p-core/host"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

func (r *REPL) handleIntrospect() error {
	// is the host introspectable
	ih, ok := r.h.(host.IntrospectableHost)
	if !ok {
		return errors.New("REPL host is not introspectable")
	}

	// create connection to WS introspection server
	addrs := ih.IntrospectionEndpoint().ListenAddrs()
	url := fmt.Sprintf("ws://%s/introspect", addrs[0])
	connection, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create connection to WS introspection server, err=%s", err))
	}
	defer connection.Close()

	// fetch & unmarshal h1 state
	var state *introspection_pb.State
	if err := connection.WriteMessage(websocket.TextMessage, []byte("trigger fetch")); err != nil {
		return errors.New(fmt.Sprintf("failed to send request to WS introspection server, err=%s", err))
	}

	// read response & unmarshal
	_, msg, err := connection.ReadMessage()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to read response from WS introspection server, err=%s", err))
	}
	state = &introspection_pb.State{}
	if err := proto.Unmarshal(msg, state); err != nil {
		return errors.New(fmt.Sprintf("failed to unmarshal response read from WS introspection server, err=%s", err))
	}

	fmt.Printf("\n -----Host Introspection Result:-----\n\n %s", proto.MarshalTextString(state))

	return err
}
