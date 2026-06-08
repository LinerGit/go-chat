package hub

import (
	"log/slog"
	"sync"
)

// Hub maintains active clients and broadcasts messages to them.
type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	log        *slog.Logger
}

func New(log *slog.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		log:        log,
	}
}

// Run starts the hub event loop. Must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()
			h.log.Info("client connected",
				"user_id", client.userID,
				"username", client.username,
				"total", h.Len(),
			)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			h.log.Info("client disconnected",
				"user_id", client.userID,
				"username", client.username,
				"total", h.Len(),
			)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// client's buffer full — drop and unregister
					h.log.Warn(
						"client buffer overflow",
						"user_id", client.userID,
					)
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					close(client.send)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(c *Client)   { h.register <- c }
func (h *Hub) Unregister(c *Client) { h.unregister <- c }
func (h *Hub) Broadcast(msg []byte) { h.broadcast <- msg }

func (h *Hub) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
