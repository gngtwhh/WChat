package handler

import (
    "wchat/internal/websocket"

    "github.com/gin-gonic/gin"
)

type WebsocketHandler struct {
    svc *websocket.WebsocketService
}

func NewWebsocketHandler(svc *websocket.WebsocketService) *WebsocketHandler {
    return &WebsocketHandler{svc: svc}
}

func (h WebsocketHandler) WsHandler(context *gin.Context) {

}
