package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/repository"
)

var (
	ErrRSVPSessionNotFound = errors.New("session not found")
	ErrRSVPSessionCanceled = errors.New("cannot RSVP to a canceled session")
	ErrRSVPSessionPast     = errors.New("cannot RSVP to a past session")
	ErrRSVPDuplicate       = errors.New("you have already RSVPed to this session")
	ErrRSVPNotFound        = errors.New("you have not RSVPed to this session")
	ErrRSVPSessionFull     = errors.New("this session is full")
)

// RSVPRepo defines the data access interface for RSVPs.
type RSVPRepo interface {
	CreateAtomic(ctx context.Context, sessionID, memberID uuid.UUID) (*repository.RSVPTxResult, error)
	Delete(ctx context.Context, sessionID, memberID uuid.UUID) (*model.SpaceSession, error)
	ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPWithMember, error)
}

// MemberGetter fetches a member by ID (for notification context).
type MemberGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error)
}

type RSVPService struct {
	repo     RSVPRepo
	members  MemberGetter
	notifier Notifier
}

func NewRSVPService(repo RSVPRepo, members MemberGetter, notifier Notifier) *RSVPService {
	return &RSVPService{repo: repo, members: members, notifier: notifier}
}

// Create creates an RSVP for the given session and member.
func (s *RSVPService) Create(ctx context.Context, sessionID, memberID uuid.UUID) (*model.RSVP, error) {
	result, err := s.repo.CreateAtomic(ctx, sessionID, memberID)
	if err != nil {
		return nil, mapRepoError(err)
	}
	if result == nil {
		return nil, ErrRSVPSessionNotFound
	}

	// Fire notification after transaction commit (fire-and-forget)
	member, _ := s.members.GetByID(ctx, memberID)
	if member != nil {
		go s.notifier.MemberRSVPed(result.Session, member)
	}

	return result.RSVP, nil
}

// Cancel removes an RSVP for the given session and member.
func (s *RSVPService) Cancel(ctx context.Context, sessionID, memberID uuid.UUID) error {
	session, err := s.repo.Delete(ctx, sessionID, memberID)
	if err != nil {
		return mapRepoError(err)
	}
	if session == nil {
		return ErrRSVPSessionNotFound
	}

	// Fire notification after successful delete (fire-and-forget)
	member, _ := s.members.GetByID(ctx, memberID)
	if member != nil {
		go s.notifier.MemberCanceledRSVP(session, member)
	}

	return nil
}

// ListBySession returns the guest list for a session.
func (s *RSVPService) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPWithMember, error) {
	rsvps, err := s.repo.ListBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("listing rsvps: %w", err)
	}
	if rsvps == nil {
		return nil, ErrRSVPSessionNotFound
	}

	return rsvps, nil
}

// mapRepoError maps repository-level sentinel errors to service-level sentinel errors.
func mapRepoError(err error) error {
	switch {
	case errors.Is(err, repository.ErrRSVPSessionCanceled):
		return ErrRSVPSessionCanceled
	case errors.Is(err, repository.ErrRSVPSessionPast):
		return ErrRSVPSessionPast
	case errors.Is(err, repository.ErrRSVPDuplicate):
		return ErrRSVPDuplicate
	case errors.Is(err, repository.ErrRSVPNotFound):
		return ErrRSVPNotFound
	case errors.Is(err, repository.ErrRSVPSessionFull):
		return ErrRSVPSessionFull
	default:
		return err
	}
}
