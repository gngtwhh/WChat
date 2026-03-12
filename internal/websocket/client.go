package websocket

import (
    "time"
    "wchat/pkg/zlog"

    "github.com/gorilla/websocket"
    "go.uber.org/zap"
)

const (
    writeWait      = 10 * time.Second    // 允许向客户端写入数据的最长时间
    pongWait       = 60 * time.Second    // 允许读取客户端下一个 Pong 响应的最长时间
    pingPeriod     = (pongWait * 9) / 10 // 向客户端发送 Ping 的周期 (必须小于 pongWait)
    maxMessageSize = 4096                // 允许读取的最大消息体积 (防恶意大包攻击)
)

type Client struct {
    Hub  *Hub
    Conn *websocket.Conn
    Send chan []byte
    Uuid string
}

// ReadPump
func (c *Client) ReadPump() {
    defer func() {
        c.Hub.Unregister <- c
        c.Conn.Close()
    }()

    c.Conn.SetReadLimit(maxMessageSize)
    c.Conn.SetReadDeadline(time.Now().Add(pongWait))

    // set heartbeat packet handler
    c.Conn.SetPongHandler(
        func(string) error {
            c.Conn.SetReadDeadline(time.Now().Add(pongWait))
            return nil
        },
    )

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
                zlog.Error("websocket read err", zap.Error(err))
            }
            break
        }

        // TODO: 【核心纽带】在这里将 message 交给 dispatcher (调度器)
        // 解析 WsMessage 协议，并调用对应的 Service 层逻辑！
        // dispatcher.Dispatch(c, message)
        _ = message
    }
}

func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.Conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.Send:
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                // hub closed the channel
                c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }

            w, err := c.Conn.NextWriter(websocket.TextMessage)
            if err != nil {
                return
            }
            w.Write(message)

            // TODO: 解决 NDJSON 的问题
            n := len(c.Send)
            for i := 0; i < n; i++ {
                w.Write([]byte{'\n'})
                w.Write(<-c.Send)
            }

            if err := w.Close(); err != nil {
                return
            }

        case <-ticker.C:
            // Ping heartbeat packet
            c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
