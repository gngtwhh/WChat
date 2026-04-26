package websocket

import (
	"context"
	"encoding/json"

	ws "wchat/internal/network/websocket"
	"wchat/internal/types"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

func (h *CommandHandler) handleMessageRecall(ctx ws.ConnectionContext, seq string, data json.RawMessage) {
	var req types.WSMessageRecallReq
	if err := json.Unmarshal(data, &req); err != nil {
		zlog.Error("message recall payload unmarshal failed", zap.Error(err))
		h.replyFailureAck(ctx, CmdMessageRecallAck, seq, "invalid message recall payload")
		return
	}
	if req.MsgUuid == "" {
		h.replyFailureAck(ctx, CmdMessageRecallAck, seq, "invalid message recall payload")
		return
	}

	result, err := h.messageService.RecallMessage(context.Background(), ctx.UserID, req.MsgUuid)
	if err != nil {
		zlog.Error("failed to recall message", zap.Error(err))
		h.replyFailureAck(ctx, CmdMessageRecallAck, seq, err.Error())
		return
	}

	h.replyAck(
		ctx,
		CmdMessageRecallAck,
		seq,
		types.WSMessageRecallAck{
			Status:        1,
			MsgUuid:       result.MsgUUID,
			MessageStatus: result.MessageStatus,
		},
	)

	pushed := make(map[string]struct{}, len(result.RecipientIDs))
	for _, recipientID := range result.RecipientIDs {
		if recipientID == "" || recipientID == ctx.UserID {
			continue
		}
		if _, ok := pushed[recipientID]; ok {
			continue
		}
		pushed[recipientID] = struct{}{}

		h.pusher.PushToUser(
			recipientID,
			ws.OutboundMessage{
				Cmd: CmdMessageRecallPush,
				Data: types.WSMessageRecallPush{
					MsgUuid:       result.MsgUUID,
					MessageStatus: result.MessageStatus,
				},
			},
		)
	}
}
