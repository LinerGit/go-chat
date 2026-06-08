package service

import (
	"context"
	"errors"
	"strings"

	"github.com/LinerGit/go-chat/internal/model"
	"github.com/LinerGit/go-chat/internal/repository"
)

var (
	ErrEmptyMessage    = errors.New("message cannot be empty")
	ErrMessageTooLong  = errors.New("message exceeds maximum length")
	ErrInvalidUsername = errors.New("invalid username")
)

const (
	MaxMessageLength = 1000
)

type ChatService interface {
	SaveMessage(
		ctx context.Context,
		userID int64,
		username string,
		content string,
	) (model.Message, error)

	GetHistory(
		ctx context.Context,
		limit int32,
	) ([]model.Message, error)
}

type chatService struct {
	repo      repository.MessageRepository
	maxMsgLen int
}

func NewChatService(
	repo repository.MessageRepository,
	maxMsgLen int,
) ChatService {

	return &chatService{
		repo:      repo,
		maxMsgLen: maxMsgLen,
	}
}

func (s *chatService) SaveMessage(
	ctx context.Context,
	userID int64,
	username string,
	content string,
) (model.Message, error) {

	content = strings.TrimSpace(content)
	username = strings.TrimSpace(username)

	if content == "" {
		return model.Message{}, ErrEmptyMessage
	}

	if len(content) > MaxMessageLength {
		return model.Message{}, ErrMessageTooLong
	}

	if username == "" {
		return model.Message{}, ErrInvalidUsername
	}

	msg, err := s.repo.CreateMessage(
		ctx,
		userID,
		username,
		content,
	)

	if err != nil {
		return model.Message{}, err
	}

	return msg, nil
}

func (s *chatService) GetHistory(
	ctx context.Context,
	limit int32,
) ([]model.Message, error) {

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	return s.repo.GetLastMessages(
		ctx,
		limit,
	)
}
