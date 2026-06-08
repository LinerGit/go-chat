package model

import "github.com/jackc/pgx/v5/pgtype"

type Message struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"user_id"`
	Username  string           `json:"username"`
	Content   string           `json:"content"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
}
