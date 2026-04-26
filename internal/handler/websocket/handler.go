package websocket

import (
	"encoding/json"

	ws "wchat/internal/network/websocket"
	"wchat/internal/service"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

type errorAck struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// CommandHandler handles business-level WebSocket commands routed from the gateway.
type CommandHandler struct {
	messageService *service.MessageService
	pusher         ws.Pusher
}

func NewCommandHandler(messageService *service.MessageService) *CommandHandler {
	return &CommandHandler{
		messageService: messageService,
	}
}

// SetPusher injects the pusher used to reply to the current connection and push to recipients.
func (h *CommandHandler) SetPusher(pusher ws.Pusher) {
	h.pusher = pusher
}

// HandleConnectionMessage implements ws.InboundHandler.
// The gateway dispatcher calls this method for commands that are not handled by
// transport-level protocol handlers, then this method matches cmd and delegates
// to the corresponding business command handler.
func (h *CommandHandler) HandleConnectionMessage(
	ctx ws.ConnectionContext, seq string, cmd int, data json.RawMessage,
) {
	switch cmd {
	case CmdChatUp:
		h.handleChat(ctx, seq, data)
	case CmdMessageListUp:
		h.handleMessageList(ctx, seq, data)
	case CmdMessageRecallUp:
		h.handleMessageRecall(ctx, seq, data)
	default:
		zlog.Warn("unsupported websocket command", zap.Int("cmd", cmd))
	}
}

func (h *CommandHandler) replyAck(ctx ws.ConnectionContext, cmd int, seq string, data any) {
	h.pusher.Reply(
		ctx,
		ws.OutboundMessage{
			Cmd:  cmd,
			Seq:  seq,
			Data: data,
		},
	)
}

func (h *CommandHandler) replyFailureAck(ctx ws.ConnectionContext, cmd int, seq string, msg string) {
	h.replyAck(
		ctx,
		cmd,
		seq,
		errorAck{
			Status:  0,
			Message: msg,
		},
	)
}
