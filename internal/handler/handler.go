package handler

// App contains all handlers
type App struct {
    Auth        *AuthHandler
    User        *UserHandler
    Contact     *ContactHandler
    Group       *GroupHandler
    Application *ApplicationHandler
    Session     *SessionHandler
    Message     *MessageHandler
    WebSocket   *WebsocketHandler
}
