package model

import db "github.com/LinerGit/go-chat/internal/repository/db"

func FromDBMessage(msg db.Message) Message {
	return Message{
		ID:        msg.ID,
		UserID:    msg.UserID,
		Username:  msg.Username,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}
}

func FromDBMessages(messages []db.Message) []Message {
	result := make([]Message, 0, len(messages))

	for _, msg := range messages {
		result = append(result, FromDBMessage(msg))
	}

	return result
}
