package repository

import (
	"context"

	"github.com/LinerGit/chat-service/internal/model"
	db "github.com/LinerGit/chat-service/internal/repository/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository interface {
	Create(ctx context.Context, userID int64, username, content string) (model.Message, error)
	GetLast(ctx context.Context, limit int32) ([]model.Message, error)
	GetByUserID(ctx context.Context, userID int64, limit int32) ([]model.Message, error)
}

type messageRepository struct {
	q *db.Queries
}

func NewMessageRepository(pool *pgxpool.Pool) MessageRepository {
	return &messageRepository{q: db.New(pool)}
}

func (r *messageRepository) Create(ctx context.Context, userID int64, username, content string) (model.Message, error) {
	row, err := r.q.CreateMessage(ctx, db.CreateMessageParams{
		UserID:   userID,
		Username: username,
		Content:  content,
	})
	if err != nil {
		return model.Message{}, err
	}
	return model.Message{
		ID:        row.ID,
		UserID:    row.UserID,
		Username:  row.Username,
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
	}, nil
}

func (r *messageRepository) GetLast(ctx context.Context, limit int32) ([]model.Message, error) {
	rows, err := r.q.GetLastMessages(ctx, limit)
	if err != nil {
		return nil, err
	}
	return mapMessages(rows), nil
}

func (r *messageRepository) GetByUserID(ctx context.Context, userID int64, limit int32) ([]model.Message, error) {
	rows, err := r.q.GetMessagesByUserID(ctx, db.GetMessagesByUserIDParams{
		UserID: userID,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	return mapMessages(rows), nil
}

func mapMessages(rows []db.Message) []model.Message {
	msgs := make([]model.Message, len(rows))
	for i, r := range rows {
		msgs[i] = model.Message{
			ID:        r.ID,
			UserID:    r.UserID,
			Username:  r.Username,
			Content:   r.Content,
			CreatedAt: r.CreatedAt,
		}
	}
	return msgs
}
