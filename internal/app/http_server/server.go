package http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
	"wchat/internal/cache"
	"wchat/internal/config"
	"wchat/internal/handler"
	"wchat/internal/handler/restful"
	"wchat/internal/handler/websocket"
	wshandler "wchat/internal/handler/websocket"
	ws "wchat/internal/network/websocket"
	"wchat/internal/repository"
	"wchat/internal/router"
	"wchat/internal/service"
	"wchat/pkg/zlog"

	"go.uber.org/zap"
)

type Server struct {
	httpServer              *http.Server
	accountLifecycleService *service.AccountLifecycleService
	webSocketGateway        *ws.Gateway
	lifecycleCtx            context.Context
	lifecycleCancel         context.CancelFunc
	runOnce                 sync.Once
	shutdownOnce            sync.Once
}

func NewServer() (h *Server) {
	cfg := config.GetConfig()

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
	txManager := repository.NewTxManager(db)

	// ==============================
	// init Services
	// ==============================
	zlog.Info("initializing service...")
	userService := service.NewUserService(userRepo, txManager)
	groupService := service.NewGroupService(groupRepo, userRepo, contactRepo)
	applicationService := service.NewApplicationService(contactApplyRepo, userRepo, contactRepo, groupRepo)
	authService := service.NewAuthService(userRepo)
	contactService := service.NewContactService(contactRepo)
	messageService := service.NewMessageService(messageRepo, sessionRepo, contactRepo, groupRepo)
	sessionService := service.NewSessionService(sessionRepo, userRepo, groupRepo)
	accountLifecycleService := service.NewAccountLifecycleService(userRepo, txManager)

	// ==============================
	// init WebSocket gateway and handler
	// (For improvement to a distributed gateway)
	// ==============================
	zlog.Info("initializing websocket gateway and handler...")
	webSocketGateway := ws.NewGateway()
	webSocketCommandHandler := wshandler.NewCommandHandler(messageService)
	webSocketCommandHandler.SetPusher(webSocketGateway)
	webSocketGateway.SetInboundHandler(webSocketCommandHandler)

	// ==============================
	// init App instance
	// ==============================
	app := &handler.App{
		Restful: handler.Restful{
			Auth:        restful.NewAuthHandler(authService),
			User:        restful.NewUserHandler(userService),
			Contact:     restful.NewContactHandler(contactService),
			Group:       restful.NewGroupHandler(groupService),
			Application: restful.NewApplicationHandler(applicationService),
			Session:     restful.NewSessionHandler(sessionService),
			Message:     restful.NewMessageHandler(messageService),
		},
		WebSocket: handler.WebSocket{
			WS: websocket.NewWebsocketHandler(webSocketGateway),
		},
	}

	h = &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.ServerConfig.Port),
			Handler: router.LoadRouters(app, authService),
		},
		accountLifecycleService: accountLifecycleService,
		webSocketGateway:        webSocketGateway,
	}
	return
}

func (s *Server) Run() {
	s.runOnce.Do(
		func() {
			s.lifecycleCtx, s.lifecycleCancel = context.WithCancel(context.Background())
			startAccountPurgeWorker(s.lifecycleCtx, s.accountLifecycleService)
			if err := s.webSocketGateway.Start(s.lifecycleCtx); err != nil {
				zlog.Error("websocket gateway startup failed", zap.Error(err))
				return
			}
			if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				if s.lifecycleCancel != nil {
					s.lifecycleCancel()
				}
				_ = s.webSocketGateway.Shutdown(context.Background())
				zlog.Error("http_server startup failed", zap.Error(err))
			}
		},
	)
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutdownErr error
	s.shutdownOnce.Do(
		func() {
			if s.lifecycleCancel != nil {
				s.lifecycleCancel()
			}
			wsErr := s.webSocketGateway.Shutdown(ctx)
			httpErr := s.httpServer.Shutdown(ctx)
			shutdownErr = errors.Join(wsErr, httpErr)
		},
	)
	return shutdownErr
}

func startAccountPurgeWorker(ctx context.Context, svc *service.AccountLifecycleService) {
	const purgeInterval = time.Hour

	go func() {
		ticker := time.NewTicker(purgeInterval)
		defer ticker.Stop()

		runOnce := func() {
			err := svc.PurgeExpiredAccounts(ctx, time.Now())
			if err != nil {
				zlog.Error("purge expired accounts failed", zap.Error(err))
				return
			}
		}

		runOnce()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runOnce()
			}
		}
	}()
}
