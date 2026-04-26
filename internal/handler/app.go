package handler

import (
	"wchat/internal/handler/restful"
	"wchat/internal/handler/websocket"
)

// App contains all handlers
type App struct {
	Restful
	WebSocket
}

type Restful struct {
	Auth        *restful.AuthHandler
	User        *restful.UserHandler
	Contact     *restful.ContactHandler
	Group       *restful.GroupHandler
	Application *restful.ApplicationHandler
	Session     *restful.SessionHandler
}

type WebSocket struct {
	WS *websocket.WebsocketHandler
}
