package websocket

import (
	"context"
	"encoding/json"

	ws "wchat/internal/network/websocket"
	"wchat/internal/service"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

type chatRequest struct {
	SessionType int8   `json:"session_type"`
	ReceiveID   string `json:"receive_id"`
	Type        int8   `json:"type"`
	Content     string `json:"content"`
	URL         string `json:"url"`
}

type CommandHandler struct {
	messageService *service.MessageService
	pusher         ws.Pusher
}

func NewCommandHandler(messageService *service.MessageService) *CommandHandler {
	return &CommandHandler{
		messageService: messageService,
	}
}

func (h *CommandHandler) SetPusher(pusher ws.Pusher) {
	h.pusher = pusher
}

func (h *CommandHandler) HandleConnectionMessage(
	ctx ws.ConnectionContext, seq string, cmd int, data json.RawMessage,
) {
	switch cmd {
	case CmdChatUp:
		h.handleChat(ctx, seq, data)
	default:
		zlog.Warn("unsupported websocket command", zap.Int("cmd", cmd))
	}
}

func (h *CommandHandler) handleChat(ctx ws.ConnectionContext, seq string, data json.RawMessage) {
	var req chatRequest
	if err := json.Unmarshal(data, &req); err != nil {
		zlog.Error("chat payload unmarshal failed", zap.Error(err))
		h.sendErrorAck(ctx, seq, "消息格式错误")
		return
	}

	result, err := h.messageService.SendMessage(
		context.Background(),
		ctx.UserID,
		req.ReceiveID,
		req.SessionType,
		req.Type,
		req.Content,
		req.URL,
	)
	if err != nil {
		zlog.Error("failed to send message", zap.Error(err))
		h.sendErrorAck(ctx, seq, err.Error())
		return
	}

	h.pusher.Reply(
		ctx,
		ws.OutboundMessage{
			Cmd: CmdChatAck,
			Seq: seq,
			Data: map[string]any{
				"msg_uuid":   result.Message.Uuid,
				"session_id": result.Message.SessionId,
				"send_at":    result.Message.SendAt,
				"status":     1,
			},
		},
	)

	for _, recipientID := range result.RecipientIDs {
		recipientMsg := *result.Message
		if sessionID, ok := result.RecipientSessionIDs[recipientID]; ok {
			recipientMsg.SessionId = sessionID
		}

		h.pusher.PushToUser(
			recipientID,
			ws.OutboundMessage{
				Cmd:  CmdChatPush,
				Data: recipientMsg,
			},
		)
	}
}

func (h *CommandHandler) sendErrorAck(ctx ws.ConnectionContext, seq string, msg string) {
	h.pusher.Reply(
		ctx,
		ws.OutboundMessage{
			Cmd:  CmdChatAck,
			Seq:  seq,
			Data: map[string]any{"status": 0, "message": msg},
		},
	)
}
