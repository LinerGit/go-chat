package hub

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
)

const sendBufferSize = 256

// Client is a single WebSocket connection.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	userID   int64
	username string
	log      *slog.Logger

	writeWait  time.Duration
	pongWait   time.Duration
	pingPeriod time.Duration
	maxMsgLen  int64
}

type ClientConfig struct {
	WriteWait  time.Duration
	PongWait   time.Duration
	PingPeriod time.Duration
	MaxMsgLen  int64
}

func NewClient(
	h *Hub,
	conn *websocket.Conn,
	userID int64,
	username string,
	cfg ClientConfig,
	log *slog.Logger,
) *Client {
	return &Client{
		hub:        h,
		conn:       conn,
		send:       make(chan []byte, sendBufferSize),
		userID:     userID,
		username:   username,
		log:        log,
		writeWait:  cfg.WriteWait,
		pongWait:   cfg.PongWait,
		pingPeriod: cfg.PingPeriod,
		maxMsgLen:  cfg.MaxMsgLen,
	}
}

type incomingMessage struct {
	Content string `json:"content"`
}

type OutgoingMessage struct {
	Type       string           `json:"type"`
	FromUserID int64            `json:"from_user_id"`
	Username   string           `json:"username"`
	Content    string           `json:"content"`
	Timestamp  pgtype.Timestamp `json:"timestamp"`
}

// ReadPump reads messages from the WebSocket connection.
// Must run in its own goroutine.
func (c *Client) ReadPump(onMessage func(userID int64, username, content string)) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(c.maxMsgLen)
	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				c.log.Warn("ws read error", "user_id", c.userID, "error", err)
			}
			return
		}

		var msg incomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil || msg.Content == "" {
			continue
		}

		onMessage(c.userID, c.username, msg.Content)
	}
}

// WritePump writes messages from the send channel to the WebSocket connection.
// Must run in its own goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				c.log.Warn("ws write error", "user_id", c.userID, "error", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(c.writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
