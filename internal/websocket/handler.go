package websocket

import (
	"context"
	"encoding/json"

	"wchat/internal/service"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

// Handler processes a specific WebSocket command.
type Handler interface {
	Handle(client *Client, seq string, data json.RawMessage)
}

// PingHandler responds to client heartbeat pings with a pong.
type PingHandler struct{}

func (h *PingHandler) Handle(client *Client, seq string, _ json.RawMessage) {
	sendToClient(client, wsMessage{Cmd: CmdPong, Seq: seq})
}

// ChatHandler processes upstream chat messages, persists them via the service
// layer, and pushes them to the receiver through the Hub.
type ChatHandler struct {
	hub        *Hub
	msgService *service.MessageService
}

func (h *ChatHandler) Handle(client *Client, seq string, data json.RawMessage) {
	var req chatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		zlog.Error("chat payload unmarshal failed", zap.Error(err))
		sendErrorAck(client, seq, "消息格式错误")
		return
	}

	ctx := context.Background()
	result, err := h.msgService.SendMessage(
		ctx,
		client.UserID,
		req.ReceiveID,
		req.SessionType,
		req.Type,
		req.Content,
		req.URL,
	)
	if err != nil {
		zlog.Error("failed to send message", zap.Error(err))
		sendErrorAck(client, seq, err.Error())
		return
	}

	// ACK to sender
	ackData, _ := json.Marshal(
		map[string]any{
			"msg_uuid":   result.Message.Uuid,
			"session_id": result.Message.SessionId,
			"send_at":    result.Message.SendAt,
			"status":     1,
		},
	)
	sendToClient(
		client, wsMessage{
			Cmd:  CmdChatAck,
			Seq:  seq,
			Data: ackData,
		},
	)

	for _, recipientID := range result.RecipientIDs {
		recipientMsg := *result.Message
		if sessionID, ok := result.RecipientSessionIDs[recipientID]; ok {
			recipientMsg.SessionId = sessionID
		}

		pushData, _ := json.Marshal(recipientMsg)
		pushEnvelope, _ := json.Marshal(
			wsMessage{
				Cmd:  CmdChatPush,
				Data: pushData,
			},
		)
		h.hub.SendToUser(recipientID, pushEnvelope)
	}
}

// sendToClient marshals a wsMessage and sends it to the client non-blockingly.
func sendToClient(client *Client, msg wsMessage) {
	b, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case client.send <- b:
	default:
		zlog.Warn(
			"client send buffer full, message dropped",
			zap.String("user_id", client.UserID),
		)
	}
}

func sendErrorAck(client *Client, seq string, msg string) {
	data, _ := json.Marshal(map[string]any{"status": 0, "message": msg})
	sendToClient(client, wsMessage{Cmd: CmdChatAck, Seq: seq, Data: data})
}
