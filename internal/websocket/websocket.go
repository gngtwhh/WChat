package websocket

import "wchat/internal/service"

type WebsocketService struct {
    messageSvc *service.MessageService
    sessionSvc *service.SessionService
}

func NewWebsocketService(
    messageSvc *service.MessageService,
    sessionSvc *service.SessionService,
) *WebsocketService {
    return &WebsocketService{
        messageSvc: messageSvc,
        sessionSvc: sessionSvc,
    }
}
