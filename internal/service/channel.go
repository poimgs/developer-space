package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

var (
	ErrChannelNotFound     = errors.New("channel not found")
	ErrCannotDeleteSession = errors.New("cannot delete session channel")
)

type ChannelRepo interface {
	Create(ctx context.Context, name, channelType string, sessionID *uuid.UUID, createdBy uuid.UUID) (*model.Channel, error)
	List(ctx context.Context) ([]model.Channel, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Channel, error)
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*model.Channel, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ChannelService struct {
	repo ChannelRepo
}

func NewChannelService(repo ChannelRepo) *ChannelService {
	return &ChannelService{repo: repo}
}

func (s *ChannelService) Create(ctx context.Context, req model.CreateChannelRequest, createdBy uuid.UUID) (*model.Channel, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, &ValidationError{Details: map[string]string{"name": "required"}}
	}
	if len(name) > 100 {
		return nil, &ValidationError{Details: map[string]string{"name": "must be 100 characters or less"}}
	}

	channel, err := s.repo.Create(ctx, name, "general", nil, createdBy)
	if err != nil {
		return nil, fmt.Errorf("creating channel: %w", err)
	}
	return channel, nil
}

func (s *ChannelService) CreateForSession(ctx context.Context, sessionTitle string, sessionID uuid.UUID, createdBy uuid.UUID) (*model.Channel, error) {
	channel, err := s.repo.Create(ctx, sessionTitle, "session", &sessionID, createdBy)
	if err != nil {
		return nil, fmt.Errorf("creating session channel: %w", err)
	}
	return channel, nil
}

func (s *ChannelService) List(ctx context.Context) ([]model.Channel, error) {
	return s.repo.List(ctx)
}

func (s *ChannelService) GetByID(ctx context.Context, id uuid.UUID) (*model.Channel, error) {
	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, ErrChannelNotFound
	}
	return channel, nil
}

func (s *ChannelService) Delete(ctx context.Context, id uuid.UUID) error {
	channel, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("getting channel: %w", err)
	}
	if channel == nil {
		return ErrChannelNotFound
	}
	if channel.Type == "session" {
		return ErrCannotDeleteSession
	}
	return s.repo.Delete(ctx, id)
}
