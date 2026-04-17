package websocket

import (
	"encoding/json"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

// WebSocket command constants.
const (
	CmdPing = 1001
	CmdPong = 1002
)

// wsMessage is the standard envelope for all WebSocket messages.
type wsMessage struct {
	Cmd  int             `json:"cmd"`
	Seq  string          `json:"seq"`
	Data json.RawMessage `json:"data"`
}

// ProtocolHandler processes transport-level WebSocket commands inside the gateway.
type ProtocolHandler interface {
	HandleConnectionMessage(ctx ConnectionContext, seq string, data json.RawMessage)
}

// Dispatcher parses incoming WebSocket messages and routes them to registered handlers.
type Dispatcher struct {
	protocolHandlers map[int]ProtocolHandler
	inbound          InboundHandler
}

func newDispatcher(inbound InboundHandler, protocolHandlers map[int]ProtocolHandler) *Dispatcher {
	return &Dispatcher{inbound: inbound, protocolHandlers: protocolHandlers}
}

// Dispatch unmarshals a raw message and delegates to the matching handler.
func (d *Dispatcher) Dispatch(ctx ConnectionContext, raw []byte) {
	var msg wsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		zlog.Error("websocket message unmarshal failed", zap.Error(err))
		return
	}

	h, ok := d.protocolHandlers[msg.Cmd]
	if !ok {
		if d.inbound == nil {
			zlog.Warn("unsupported websocket command", zap.Int("cmd", msg.Cmd))
			return
		}
		d.inbound.HandleConnectionMessage(ctx, msg.Seq, msg.Cmd, msg.Data)
		return
	}
	h.HandleConnectionMessage(ctx, msg.Seq, msg.Data)
}

func (d *Dispatcher) SetInboundHandler(inbound InboundHandler) {
	d.inbound = inbound
}
