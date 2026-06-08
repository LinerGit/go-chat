package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/LinerGit/chat-service/internal/hub"
	"github.com/LinerGit/chat-service/internal/middleware"
	"github.com/LinerGit/chat-service/internal/service"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	chat      service.ChatService
	hub       *hub.Hub
	upgrader  websocket.Upgrader
	clientCfg hub.ClientConfig
	log       *slog.Logger
}

func NewChatHandler(
	chat service.ChatService,
	h *hub.Hub,
	clientCfg hub.ClientConfig,
	log *slog.Logger,
) *ChatHandler {
	return &ChatHandler{
		chat: chat,
		hub:  h,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: restrict to allowed origins in production
				return true
			},
		},
		clientCfg: clientCfg,
		log:       log,
	}
}

// ServeWS godoc
//
// @Summary      WebSocket chat endpoint
// @Description  Upgrades to WebSocket. JWT required via ?token= query param or Authorization header.
// @Tags         ws
// @Router       /ws/chat [get]
func (h *ChatHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromCtx(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("ws upgrade failed", "error", err)
		return
	}

	client := hub.NewClient(
		h.hub,
		conn,
		claims.UserID,
		claims.Username,
		h.clientCfg,
		h.log,
	)

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(h.onMessage)
}

// onMessage is the ReadPump callback — runs in a goroutine per client.
func (h *ChatHandler) onMessage(userID int64, username, content string) {
	ctx := context.Background()

	msg, err := h.chat.SaveMessage(ctx, userID, username, content)
	if err != nil {
		h.log.Warn("save message failed", "error", err, "user_id", userID)
		return
	}

	payload := hub.OutgoingMessage{
		Type:       "message",
		FromUserID: msg.UserID,
		Username:   msg.Username,
		Content:    msg.Content,
		Timestamp:  msg.CreatedAt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		h.log.Error("marshal message failed", "error", err)
		return
	}

	h.hub.Broadcast(data)
}

// GetHistory godoc
//
// @Summary      Get message history
// @Description  Returns last 50 messages ordered by newest first
// @Tags         messages
// @Produce      json
// @Success      200  {array}   model.Message
// @Failure      500
// @Router       /messages [get]
func (h *ChatHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	msgs, err := h.chat.GetHistory(r.Context(), 50)
	if err != nil {
		h.log.Error("get history failed", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	render.JSON(w, r, msgs)
}

// Health godoc
//
// @Summary      Health check
// @Tags         system
// @Produce      json
// @Success      200
// @Router       /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}
