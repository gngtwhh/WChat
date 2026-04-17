package websocket

import "encoding/json"

// PingHandler responds to client heartbeat pings with a pong.
type PingHandler struct {
	pusher Pusher
}

func (h *PingHandler) HandleConnectionMessage(ctx ConnectionContext, seq string, _ json.RawMessage) {
	h.pusher.Reply(ctx, OutboundMessage{Cmd: CmdPong, Seq: seq})
}

func marshalOutboundMessage(msg OutboundMessage) ([]byte, error) {
	payload := wsMessage{
		Cmd: msg.Cmd,
		Seq: msg.Seq,
	}
	if msg.Data != nil {
		data, err := json.Marshal(msg.Data)
		if err != nil {
			return nil, err
		}
		payload.Data = data
	}
	return json.Marshal(payload)
}
