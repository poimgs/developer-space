package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

type MessageRepo interface {
	Create(ctx context.Context, channelID, memberID uuid.UUID, content string) (*model.MessageWithAuthor, error)
	ListByChannel(ctx context.Context, channelID uuid.UUID, cursor *time.Time, limit int) (*model.MessagePage, error)
}

type MessageService struct {
	repo        MessageRepo
	channelRepo ChannelRepo
}

func NewMessageService(repo MessageRepo, channelRepo ChannelRepo) *MessageService {
	return &MessageService{repo: repo, channelRepo: channelRepo}
}

func (s *MessageService) Send(ctx context.Context, channelID, memberID uuid.UUID, content string) (*model.MessageWithAuthor, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, &ValidationError{Details: map[string]string{"content": "required"}}
	}
	if len(content) > 4000 {
		return nil, &ValidationError{Details: map[string]string{"content": "must be 4000 characters or less"}}
	}

	// Verify channel exists
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("checking channel: %w", err)
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	msg, err := s.repo.Create(ctx, channelID, memberID, content)
	if err != nil {
		return nil, fmt.Errorf("sending message: %w", err)
	}
	return msg, nil
}

func (s *MessageService) ListHistory(ctx context.Context, channelID uuid.UUID, cursor *time.Time, limit int) (*model.MessagePage, error) {
	// Verify channel exists
	channel, err := s.channelRepo.GetByID(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("checking channel: %w", err)
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}

	return s.repo.ListByChannel(ctx, channelID, cursor, limit)
}
