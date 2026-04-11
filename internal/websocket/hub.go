package websocket

import "sync"

// Hub maintains the set of active clients and routes messages to online users.
type Hub struct {
	clients    map[string]map[*Client]bool // UserID -> connections, supports multi-device login
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run listens for register/unregister events and maintains the client map.
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				if _, exists := conns[client]; exists {
					delete(conns, client)
					close(client.send)
					if len(conns) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser delivers a message to all active connections of the given user.
// Non-blocking: drops the message if a client's send buffer is full.
func (h *Hub) SendToUser(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range conns {
		select {
		case client.send <- message:
		default:
		}
	}
}

// Broadcast delivers a message to all connected clients.
func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, conns := range h.clients {
		for client := range conns {
			select {
			case client.send <- message:
			default:
			}
		}
	}
}
