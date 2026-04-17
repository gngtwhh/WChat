package websocket

import (
	"context"
	"sync"
)

// Hub maintains the set of active clients and routes messages to online users.
type Hub struct {
	clients     map[string]map[*Client]bool // UserID -> connections, supports multi-device login
	connections map[string]*Client
	mu          sync.RWMutex
	register    chan *Client
	unregister  chan *Client
	running     bool
	done        chan struct{}
	cancel      context.CancelFunc
	startOnce   sync.Once
	stopOnce    sync.Once
}

func newHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		connections: make(map[string]*Client),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		done:        make(chan struct{}),
	}
}

// run listens for register/unregister events and maintains the client map.
func (h *Hub) run(ctx context.Context) {
	defer close(h.done)
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.connections[client.ID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			h.removeClientLocked(client)
			h.mu.Unlock()

		case <-ctx.Done():
			h.shutdownClients()
			h.mu.Lock()
			h.running = false
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) Start(ctx context.Context) {
	h.startOnce.Do(
		func() {
			runCtx, cancel := context.WithCancel(ctx)
			h.mu.Lock()
			h.running = true
			h.cancel = cancel
			h.mu.Unlock()
			go h.run(runCtx)
		},
	)
}

func (h *Hub) Shutdown(ctx context.Context) error {
	if h.cancel == nil {
		return nil
	}

	h.stopOnce.Do(
		func() {
			h.cancel()
		},
	)

	select {
	case <-h.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *Hub) RegisterClient(client *Client) bool {
	h.mu.RLock()
	running := h.running
	h.mu.RUnlock()
	if !running {
		client.conn.Close()
		close(client.send)
		return false
	}

	select {
	case h.register <- client:
		return true
	case <-h.done:
		client.conn.Close()
		close(client.send)
		return false
	}
}

func (h *Hub) UnregisterClient(client *Client) {
	h.mu.RLock()
	running := h.running
	h.mu.RUnlock()
	if !running {
		h.removeClient(client)
		return
	}

	select {
	case h.unregister <- client:
	case <-h.done:
		h.removeClient(client)
	}
}

// SendToConnection delivers a message to a specific active connection.
func (h *Hub) SendToConnection(connectionID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, ok := h.connections[connectionID]
	if !ok {
		return
	}

	select {
	case client.send <- message:
	default:
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

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeClientLocked(client)
}

func (h *Hub) removeClientLocked(client *Client) {
	if conns, ok := h.clients[client.UserID]; ok {
		if _, exists := conns[client]; exists {
			delete(conns, client)
			delete(h.connections, client.ID)
			close(client.send)
			if len(conns) == 0 {
				delete(h.clients, client.UserID)
			}
		}
	}
}

func (h *Hub) shutdownClients() {
	h.mu.Lock()
	clients := make([]*Client, 0, len(h.connections))
	for _, client := range h.connections {
		clients = append(clients, client)
		close(client.send)
	}
	h.clients = make(map[string]map[*Client]bool)
	h.connections = make(map[string]*Client)
	h.mu.Unlock()

	for _, client := range clients {
		client.conn.Close()
	}
}
