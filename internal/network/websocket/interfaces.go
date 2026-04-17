package websocket

import "encoding/json"

// ConnectionContext identifies the active WebSocket connection and user.
type ConnectionContext struct {
	ConnectionID string
	UserID       string
}

// InboundHandler handles parsed WebSocket commands outside the gateway runtime.
type InboundHandler interface {
	HandleConnectionMessage(ctx ConnectionContext, seq string, cmd int, data json.RawMessage)
}

// OutboundMessage is the public message envelope used by the gateway push APIs.
type OutboundMessage struct {
	Cmd  int    `json:"cmd"`
	Seq  string `json:"seq,omitempty"`
	Data any    `json:"data,omitempty"`
}

// Pusher sends replies to the active connection or pushes to online users.
type Pusher interface {
	Reply(ctx ConnectionContext, msg OutboundMessage)
	PushToUser(userID string, msg OutboundMessage)
}
