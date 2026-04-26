package websocket

import (
	"context"
	"encoding/json"

	ws "wchat/internal/network/websocket"
	"wchat/internal/types"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

// handleChat handles CmdChatUp by sending the message, acknowledging the sender,
// and fan-out pushing the final message to all recipients.
func (h *CommandHandler) handleChat(ctx ws.ConnectionContext, seq string, data json.RawMessage) {
	var req types.WSChatReq
	if err := json.Unmarshal(data, &req); err != nil {
		zlog.Error("chat payload unmarshal failed", zap.Error(err))
		h.replyFailureAck(ctx, CmdChatAck, seq, "invalid chat payload")
		return
	}

	result, err := h.messageService.SendMessage(
		context.Background(),
		ctx.UserID,
		req.ReceiveId,
		req.SessionType,
		req.Type,
		req.Content,
		req.Url,
	)
	if err != nil {
		zlog.Error("failed to send message", zap.Error(err))
		h.replyFailureAck(ctx, CmdChatAck, seq, err.Error())
		return
	}

	h.replyAck(
		ctx,
		CmdChatAck,
		seq,
		map[string]any{
			"msg_uuid":   result.Message.Uuid,
			"session_id": result.Message.SessionId,
			"send_at":    result.Message.SendAt,
			"status":     1,
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
