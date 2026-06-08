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
	r.Use(corsMiddleware)

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

// corsMiddleware handles preflight OPTIONS requests and adds CORS headers to
// every response. Uses stdlib only — no extra dependency required.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
