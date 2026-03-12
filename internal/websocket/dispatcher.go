package websocket

import (
    "context"
    "encoding/json"
    "log"
    "wchat/pkg/zlog"

    "wchat/internal/service"

    "go.uber.org/zap"
)

const (
    CmdPing        = 1001 // 客户端发来的心跳
    CmdPong        = 1002 // 服务端回复的心跳
    CmdChatUp      = 2001 // 客户端上行聊天消息
    CmdChatAck     = 2002 // 服务端回执 (告诉发送者成功了)
    CmdChatPush    = 2003 // 服务端下推新消息 (推给接收者)
    CmdSystemEvent = 3001 // 系统事件下发 (撤回、加好友等)
)

type wsMessage struct {
    Cmd  int             `json:"cmd"`
    Seq  string          `json:"seq"`
    Data json.RawMessage `json:"data"`
}

type wsChatMsg struct {
    SessionType int8   `json:"session_type"` // 0单聊，1群聊
    ReceiveId   string `json:"receive_id"`
    Type        int8   `json:"type"` // 0文本，1语音，2图片/文件
    Content     string `json:"content"`
    Url         string `json:"url"`
}

type wsEventPush struct {
    EventType  string `json:"event_type"`
    TargetUuid string `json:"target_uuid"`
    Message    string `json:"message"`
}

type Dispatcher struct {
    hub        *Hub
    msgService *service.MessageService
    // groupService *service.GroupService
}

func NewDispatcher(hub *Hub, msgService *service.MessageService) *Dispatcher {
    return &Dispatcher{
        hub:        hub,
        msgService: msgService,
    }
}

func (d *Dispatcher) Dispatch(client *Client, rawMessage []byte) {
    var wsMsg wsMessage
    if err := json.Unmarshal(rawMessage, &wsMsg); err != nil {
        zlog.Error("Websocket 消息解析失败", zap.Error(err))
        return
    }

    switch wsMsg.Cmd {
    case CmdPing:
        d.handlePing(client, wsMsg.Seq)
    case CmdChatUp:
        d.handleChatMessage(client, wsMsg)
    default:
        zlog.Error("Websocket 未知或不支持的指令 CMD", zap.Int("cmd", wsMsg.Cmd))
    }
}

func (d *Dispatcher) handlePing(client *Client, seq string) {
    d.sendToClient(
        client, wsMessage{
            Cmd: CmdPong,
            Seq: seq,
        },
    )
}

func (d *Dispatcher) handleChatMessage(client *Client, envelope wsMessage) {
    var chatReq wsChatMsg
    if err := json.Unmarshal(envelope.Data, &chatReq); err != nil {
        zlog.Error("Websocket 聊天载荷解析失败", zap.Error(err))
        return
    }

    ctx := context.Background()
    msgVO, err := d.msgService.SendMessage(
        ctx,
        client.UserID,
        chatReq.ReceiveId,
        chatReq.SessionType,
        chatReq.Type,
        chatReq.Content,
        chatReq.Url,
    )

    if err != nil {
        // 业务拒绝 (如不是好友、被拉黑)，可以下发错误指令，此处暂时打日志记录
        zlog.Error("Websocket 消息发送被 Service 拒绝", zap.Error(err))
        return
    }

    ackData, _ := json.Marshal(
        map[string]any{
            "msg_uuid": msgVO.Uuid,
            "send_at":  msgVO.SendAt,
            "status":   1,
        },
    )
    d.sendToClient(
        client, wsMessage{
            Cmd:  CmdChatAck,
            Seq:  envelope.Seq,
            Data: ackData,
        },
    )

    pushData, _ := json.Marshal(msgVO)
    pushEnvelope, _ := json.Marshal(
        wsMessage{
            Cmd:  CmdChatPush,
            Seq:  "",
            Data: pushData,
        },
    )

    // 4. 路由精准投递
    if chatReq.SessionType == 0 {
        // 单聊：推给目标用户的所有在线设备
        d.hub.SendToUser(chatReq.ReceiveId, pushEnvelope)
    } else if chatReq.SessionType == 1 {
        // 群聊：TODO 调用 d.groupService 查出所有成员 UserID，循环调用 d.hub.SendToUser 推送
        // 需要注意的是，群聊发送不要推给自己 (因为自己已经收到 ACK 了)
    }
}

// sendToClient 快捷方法：把组装好的私有信封序列化，并安全塞入通道
func (d *Dispatcher) sendToClient(client *Client, msg wsMessage) {
    b, err := json.Marshal(msg)
    if err != nil {
        return
    }

    // 非阻塞发送，保护 Hub 和整个服务不被卡死
    select {
    case client.Send <- b:
    default:
        log.Printf("[WS] 客户端 %s 的发送管道已满，消息丢弃", client.UserID)
    }
}
