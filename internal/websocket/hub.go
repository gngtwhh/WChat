package websocket

import "sync"

type Hub struct {
    clients    map[string]map[*Client]bool // map Uuid --> map[*Client]bool | 多端登录
    mu         sync.RWMutex
    Register   chan *Client // 客户端上线
    Unregister chan *Client // 客户端下线
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]map[*Client]bool),
        Register:   make(chan *Client),
        Unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.Register:
            // handle online
            h.mu.Lock()
            if h.clients[client.Uuid] == nil {
                h.clients[client.Uuid] = make(map[*Client]bool)
            }
            h.clients[client.Uuid][client] = true
            h.mu.Unlock()

        case client := <-h.Unregister:
            // handle offline
            h.mu.Lock()
            if connections, ok := h.clients[client.Uuid]; ok {
                if _, exists := connections[client]; exists {
                    delete(connections, client)
                    close(client.Send)

                    if len(connections) == 0 {
                        delete(h.clients, client.Uuid)
                    }
                }
            }
            h.mu.Unlock()
        }
    }
}

func (h *Hub) SendToUser(userID string, message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    // check online
    connections, ok := h.clients[userID]
    if !ok {
        return
    }

    for client := range connections {
        select {
        case client.Send <- message:
            // 成功塞入该客户端的发送管道
        default:
            // 如果发送管道满了 (比如客户端网络极差卡死了)
            // 为了不阻塞其他推送，这里直接走 default 丢弃，Client 的 ReadPump 会发现网络断开并触发清理
        }
    }
}

func (h *Hub) Broadcast(message []byte) {
    h.mu.RLock()
    defer h.mu.RUnlock()

    for _, connections := range h.clients {
        for client := range connections {
            select {
            case client.Send <- message:
            default:
            }
        }
    }
}
