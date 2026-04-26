package websocket

import (
	"context"
	"encoding/json"

	ws "wchat/internal/network/websocket"
	"wchat/internal/service"
	"wchat/internal/types"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

func (h *CommandHandler) handleMessageList(ctx ws.ConnectionContext, seq string, data json.RawMessage) {
	var req types.WSMessageListReq
	if err := json.Unmarshal(data, &req); err != nil {
		zlog.Error("message list payload unmarshal failed", zap.Error(err))
		h.replyFailureAck(ctx, CmdMessageListAck, seq, "invalid message list payload")
		return
	}
	if req.SessionId == "" || req.Page < 1 || req.Size < 1 || req.Size > 100 {
		h.replyFailureAck(ctx, CmdMessageListAck, seq, "invalid message list payload")
		return
	}

	total, infos, err := h.messageService.GetMessageList(context.Background(), ctx.UserID, req.SessionId, req.Page, req.Size)
	if err != nil {
		zlog.Error("failed to get message list", zap.Error(err))
		h.replyFailureAck(ctx, CmdMessageListAck, seq, err.Error())
		return
	}

	list := make([]types.MessageVO, 0, len(infos))
	for _, info := range infos {
		list = append(list, newMessageVO(info))
	}

	h.replyAck(
		ctx,
		CmdMessageListAck,
		seq,
		types.WSMessageListAck{
			Status: 1,
			Total:  total,
			List:   list,
		},
	)
}

func newMessageVO(info service.MessageInfo) types.MessageVO {
	return types.MessageVO{
		Uuid:      info.Uuid,
		SessionId: info.SessionId,
		Type:      info.Type,
		Content:   info.Content,
		Url:       info.Url,
		SendId:    info.SendId,
		ReceiveId: info.ReceiveId,
		FileType:  info.FileType,
		FileName:  info.FileName,
		FileSize:  info.FileSize,
		AVdata:    info.AVdata,
		Status:    info.Status,
		SendAt:    info.SendAt,
	}
}
