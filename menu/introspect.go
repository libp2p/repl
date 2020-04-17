package menu

import (
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p-core/host"
	introspection_pb "github.com/libp2p/go-libp2p-core/introspection/pb"

	"github.com/gorilla/websocket"
	"github.com/gogo/protobuf/proto"
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

	// first read the Runtime message
	_, bz, err := connection.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read runtime message from WS server, err%s", err)
	}
	wrapper := &introspection_pb.ProtocolDataPacket{}
	if err := proto.Unmarshal(bz, wrapper); err != nil {
		return fmt.Errorf("failed to unamrshal runtime message, err=%s", err)
	}
	fmt.Printf("\n -----Introspection Runtime Result:-----\n\n %s", proto.MarshalTextString(wrapper.GetRuntime()))

	// Then the State message
	_, bz, err = connection.ReadMessage()
	if err != nil {
		return fmt.Errorf("failed to read state message from WS server, err%s", err)
	}
	if err := proto.Unmarshal(bz, wrapper); err != nil {
		return fmt.Errorf("failed to unamrshal state message, err=%s", err)
	}
	fmt.Printf("\n -----Introspection State Result:-----\n\n %s", proto.MarshalTextString(wrapper.GetState()))

	return err
}
