package websocket

import (
	"net/http"
	"wchat/internal/service"
	"wchat/pkg/zlog"

	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: restrict origins in production
	},
}

// WebsocketService is the entry point of the WebSocket subsystem.
// It wires together the Hub, Dispatcher, and command Handlers.
type WebsocketService struct {
	hub        *Hub
	dispatcher *Dispatcher
}

// NewWebsocketService creates a fully wired WebSocket service and starts the hub.
func NewWebsocketService(
	messageSvc *service.MessageService, sessionSvc *service.SessionService,
) *WebsocketService {
	hub := newHub()
	go hub.run()

	dispatcher := newDispatcher(
		map[int]Handler{
			CmdPing:   &PingHandler{},
			CmdChatUp: &ChatHandler{hub: hub, msgService: messageSvc},
		},
	)

	_ = sessionSvc // reserved for future handlers (e.g. session sync)

	return &WebsocketService{
		hub:        hub,
		dispatcher: dispatcher,
	}
}

// ServeWS upgrades an HTTP connection to WebSocket and starts serving the client.
// It blocks until the connection is closed, so call it at the end of the HTTP handler.
func (s *WebsocketService) ServeWS(w http.ResponseWriter, r *http.Request, userID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zlog.Error("websocket upgrade failed", zap.Error(err))
		return
	}

	client := &Client{
		hub:      s.hub,
		conn:     conn,
		send:     make(chan []byte, sendBufferSize),
		UserID:   userID,
		dispatch: s.dispatcher.Dispatch,
	}

	s.hub.register <- client
	go client.writePump()
	client.readPump() // blocks until disconnect
}
