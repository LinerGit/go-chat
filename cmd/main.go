package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LinerGit/go-chat/internal/config"
	"github.com/LinerGit/go-chat/internal/database"
	"github.com/LinerGit/go-chat/internal/handler"
	"github.com/LinerGit/go-chat/internal/hub"
	authmw "github.com/LinerGit/go-chat/internal/middleware"
	"github.com/LinerGit/go-chat/internal/repository"
	"github.com/LinerGit/go-chat/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.MustLoad()

	// database
	pool, err := database.New(cfg.DBDSN)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// repositories
	msgRepo := repository.NewMessageRepository(pool)

	// services
	chatSvc := service.NewChatService(msgRepo, cfg.MaxMsgLen)

	// hub
	h := hub.New(log)
	go h.Run()

	// client config
	clientCfg := hub.ClientConfig{
		WriteWait:  cfg.WriteWait,
		PongWait:   cfg.PongWait,
		PingPeriod: cfg.PingPeriod,
		MaxMsgLen:  int64(cfg.MaxMsgLen),
	}

	// middleware & handlers
	authMw := authmw.NewAuthMiddleware(cfg.JWTSecret)
	chatHandler := handler.NewChatHandler(chatSvc, h, clientCfg, log)

	// router
	r := handler.NewRouter(chatHandler, authMw)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // 0 = no timeout for WebSocket connections
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("server stopped")
}
