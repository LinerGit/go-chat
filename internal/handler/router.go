package handler

import (
	"net/http"
	"time"

	authmw "github.com/LinerGit/go-chat/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	chatHandler *ChatHandler,
	authMw *authmw.AuthMiddleware,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public
	r.Get("/health", Health)

	// Protected REST
	r.Group(func(r chi.Router) {
		r.Use(authMw.Auth)
		r.Get("/messages", chatHandler.GetHistory)
	})

	// WebSocket (uses AuthWS — supports ?token= query param)
	r.Group(func(r chi.Router) {
		r.Use(authMw.AuthWS)
		r.Get("/ws/chat", chatHandler.ServeWS)
	})

	return r
}
