package websocket

import (
	"context"
	"net/http"
	"wchat/pkg/zlog"

	ws "github.com/gorilla/websocket"
	"github.com/rs/xid"
	"go.uber.org/zap"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict origins in production
	},
}

// Gateway is the entry point of the WebSocket subsystem.
// It wires together the Hub and Dispatcher.
type Gateway struct {
	hub        *Hub
	dispatcher *Dispatcher
}

// NewGateway creates a fully wired WebSocket gateway.
func NewGateway() *Gateway {
	hub := newHub()

	gateway := &Gateway{
		hub:        hub,
		dispatcher: nil,
	}
	gateway.dispatcher = newDispatcher(
		nil,
		map[int]ProtocolHandler{
			CmdPing: &PingHandler{pusher: gateway},
		},
	)
	return gateway
}

func (s *Gateway) Start(ctx context.Context) error {
	s.hub.Start(ctx)
	return nil
}

func (s *Gateway) Shutdown(ctx context.Context) error {
	return s.hub.Shutdown(ctx)
}

func (s *Gateway) SetInboundHandler(handler InboundHandler) {
	s.dispatcher.SetInboundHandler(handler)
}

func (s *Gateway) Reply(ctx ConnectionContext, msg OutboundMessage) {
	payload, err := marshalOutboundMessage(msg)
	if err != nil {
		return
	}
	s.hub.SendToConnection(ctx.ConnectionID, payload)
}

func (s *Gateway) PushToUser(userID string, msg OutboundMessage) {
	payload, err := marshalOutboundMessage(msg)
	if err != nil {
		return
	}
	s.hub.SendToUser(userID, payload)
}

// ServeWS upgrades an HTTP connection to WebSocket and starts serving the client.
// It blocks until the connection is closed, so call it at the end of the HTTP handler.
func (s *Gateway) ServeWS(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zlog.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		hub:      s.hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		ID:       xid.New().String(),
		UserID:   userID,
		dispatch: s.dispatcher.Dispatch,
	}

	if !s.hub.RegisterClient(client) {
		return
	}
	go client.writePump()
	client.readPump() // blocks until disconnect
}
