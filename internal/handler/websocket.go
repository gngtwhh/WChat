package handler

import (
	"wchat/internal/service"
	"wchat/internal/websocket"
	"wchat/pkg/response"

	"github.com/gin-gonic/gin"
)

type WebsocketHandler struct {
	svc     *websocket.WebsocketService
	authSvc *service.AuthService
}

func NewWebsocketHandler(svc *websocket.WebsocketService, authSvc *service.AuthService) *WebsocketHandler {
	return &WebsocketHandler{svc: svc, authSvc: authSvc}
}

// WsHandler WebSocket 连接入口
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
// @Param        token  query  string  true  "JWT Token"
// @Success      101    "协议升级成功，WebSocket 连接已建立"
// @Failure      200    {object}  response.Response{data=nil}  "Token 无效 / Token 缺失"
// @Router       /ws [get]
func (h WebsocketHandler) WsHandler(c *gin.Context) {
	token := c.Query("token")
	user, err := h.authSvc.ValidateToken(c.Request.Context(), token)
	if err != nil {
		response.FailErr(c, err)
		return
	}

	// Upgrade to WebSocket and start serving
	h.svc.ServeWS(c.Writer, c.Request, user.Uuid)
}
