package websocket

import (
	"wchat/internal/middleware"
	"wchat/internal/network/websocket"
	"wchat/pkg/errcode"
	"wchat/pkg/response"

	"github.com/gin-gonic/gin"
)

type WebsocketHandler struct {
	gateway *websocket.Gateway
}

func NewWebsocketHandler(gateway *websocket.Gateway) *WebsocketHandler {
	return &WebsocketHandler{gateway: gateway}
}

// WsUpgrade WebSocket 连接入口
// @Summary      WebSocket 连接
// @Description  将 HTTP 连接升级为 WebSocket 长连接，用于实时消息推送和收发。
//
//	客户端通过 query 参数传递 JWT Token 进行鉴权。
//	连接成功后可发送以下指令：
//	- cmd=1001 Ping 心跳（服务端回复 cmd=1002 Pong）
//	- cmd=2001 发送聊天消息（服务端回复 cmd=2002 ACK，并通过 cmd=2003 推送给接收方）
//	- cmd=3001 系统事件（服务端主动推送）
//
// @Tags         WebSocket
// @Success      101    "协议升级成功，WebSocket 连接已建立"
// @Failure      200    {object}  response.Response{data=nil}  "Token 无效 / Token 缺失"
// @Router       /ws [get]
func (h *WebsocketHandler) WsUpgrade(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Fail(c, errcode.TokenInvalid)
		return
	}

	// Upgrade to WebSocket and start serving
	h.gateway.ServeWS(c.Writer, c.Request, userID)
}
