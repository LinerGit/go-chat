package model

import "time"

const (
	MessageTypeChat   = "message"
	MessageTypeJoin   = "join"
	MessageTypeLeave  = "leave"
	MessageTypeSystem = "system"
)

type WebSocketMessage struct {
	Type      string    `json:"type"`
	UserID    int64     `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type IncomingMessage struct {
	Content string `json:"content"`
}
