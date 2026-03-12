package http_server

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "wchat/internal/cache"
    "wchat/internal/config"
    "wchat/internal/handler"
    "wchat/internal/repository"
    "wchat/internal/router"
    "wchat/internal/service"
    "wchat/internal/websocket"
    "wchat/pkg/zlog"

    "go.uber.org/zap"
)

type Server struct {
    httpServer *http.Server
    // chatWebSocketServer *ChatWebSocketServer
}

func NewServer() (h *Server) {
    cfg := config.GetConfig()

    // ==============================
    // utils
    // ==============================
    // jwt
    // if err := utils.InitJwt(cfg.ServerConfig.JwtSecret); err != nil {
    //     zlog.Error("failed to init jwt pkg", zap.Error(err))
    //     panic(err)
    // }

    // sensitive words filter
    // file, err := os.Open(cfg.ServerConfig.SensitiveWordsFile)
    // if err != nil {
    //     zlog.Error("failed to load sensitive words file", zap.Error(err))
    //     panic(err)
    // }
    // defer file.Close()
    //
    // var words []string
    // scanner := bufio.NewScanner(file)
    // for scanner.Scan() {
    //     word := strings.TrimSpace(scanner.Text())
    //     if word != "" {
    //         words = append(words, word)
    //     }
    // }
    // acFilter := sensitive.NewACFilter()
    // acFilter.Build(words)

    // ==============================
    // init redis cache
    // ==============================
    if err := cache.InitRedis(cfg.RedisConfig.Host, cfg.RedisConfig.Port, cfg.RedisConfig.Password); err != nil {
        zlog.Error("init cache failed", zap.Error(err))
        panic(err)
    }

    // ==============================
    // init repository --- MySQL
    // ==============================
    zlog.Info("initializing database...")
    db, err := repository.InitDB()
    if err != nil {
        zlog.Error("failed to connect database", zap.Error(err))
        panic(err)
    }
    userRepo := repository.NewUserRepo(db)
    groupRepo := repository.NewGroupRepo(db)
    contactRepo := repository.NewContactRepo(db)
    contactApplyRepo := repository.NewContactApplyRepo(db)
    sessionRepo := repository.NewSessionRepo(db)
    messageRepo := repository.NewMessageRepo(db)

    // ==============================
    // init Services
    // ==============================
    zlog.Info("initializing service...")
    userService := service.NewUserService(userRepo)
    groupService := service.NewGroupService(groupRepo, userRepo, contactRepo)
    applicationService := service.NewApplicationService(contactApplyRepo, userRepo, contactRepo)
    authService := service.NewAuthService(userRepo)
    contactService := service.NewContactService(contactRepo)
    messageService := service.NewMessageService(messageRepo, sessionRepo)
    sessionService := service.NewSessionService(sessionRepo, userRepo, groupRepo)
    webSocketService := websocket.NewWebsocketService(messageService, sessionService)

    // ==============================
    // init handler
    // ==============================
    app := &handler.App{
        Auth:        handler.NewAuthHandler(authService),
        User:        handler.NewUserHandler(userService),
        Contact:     handler.NewContactHandler(contactService),
        Group:       handler.NewGroupHandler(groupService),
        Application: handler.NewApplicationHandler(applicationService),
        Session:     handler.NewSessionHandler(sessionService),
        Message:     handler.NewMessageHandler(messageService),
        WebSocket:   handler.NewWebsocketHandler(webSocketService),
    }

    // html template pre-compile
    // zlog.Info("pre-compiling html templates...")
    // tmpls := loadTmlps()
    // render.Init(tmpls, "layout")

    h = &Server{
        httpServer: &http.Server{
            Addr:    fmt.Sprintf(":%d", cfg.ServerConfig.Port),
            Handler: router.LoadRouters(app),
        },
    }
    return
}

func (s *Server) Run() {
    if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        zlog.Error("http_server startup failed", zap.Error(err))
    }
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.httpServer.Shutdown(ctx)
}

// func loadTmlps() map[string]*template.Template {
//     tmpls := make(map[string]*template.Template)
//
//     baseDir := cfg.App.TemplateDir
//     layout := baseDir + "layout/layout.html"
//
//     tmpls["index"] = template.Must(template.ParseFiles(layout, baseDir+"index.html"))
//     tmpls["admin"] = template.Must(template.ParseFiles(layout, baseDir+"admin.html"))
//     tmpls["article"] = template.Must(template.ParseFiles(layout, baseDir+"article.html"))
//     // tmpls["layout"] = template.Must(template.ParseFiles("web/templates/layout.html"))
//     return tmpls
// }
