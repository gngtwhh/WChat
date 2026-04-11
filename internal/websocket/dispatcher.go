package websocket

import (
	"encoding/json"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

// WebSocket command constants.
const (
	CmdPing        = 1001
	CmdPong        = 1002
	CmdChatUp      = 2001
	CmdChatAck     = 2002
	CmdChatPush    = 2003
	CmdSystemEvent = 3001
)

// wsMessage is the standard envelope for all WebSocket messages.
type wsMessage struct {
	Cmd  int             `json:"cmd"`
	Seq  string          `json:"seq"`
	Data json.RawMessage `json:"data"`
}

// chatRequest represents the payload of a CmdChatUp message.
type chatRequest struct {
	SessionType int8   `json:"session_type"`
	ReceiveID   string `json:"receive_id"`
	Type        int8   `json:"type"`
	Content     string `json:"content"`
	URL         string `json:"url"`
}

// eventPush represents a system event notification payload.
type eventPush struct {
	EventType  string `json:"event_type"`
	TargetUUID string `json:"target_uuid"`
	Message    string `json:"message"`
}

// Dispatcher parses incoming WebSocket messages and routes them to registered handlers.
type Dispatcher struct {
	handlers map[int]Handler
}

func newDispatcher(handlers map[int]Handler) *Dispatcher {
	return &Dispatcher{handlers: handlers}
}

// Dispatch unmarshals a raw message and delegates to the matching handler.
func (d *Dispatcher) Dispatch(client *Client, raw []byte) {
	var msg wsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		zlog.Error("websocket message unmarshal failed", zap.Error(err))
		return
	}

	h, ok := d.handlers[msg.Cmd]
	if !ok {
		zlog.Warn("unsupported websocket command", zap.Int("cmd", msg.Cmd))
		return
	}
	h.Handle(client, msg.Seq, msg.Data)
}
