package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrHasRSVPs       = errors.New("member has existing RSVPs")
)

// MemberRepo defines the data access interface for members.
type MemberRepo interface {
	Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error)
	List(ctx context.Context, activeFilter string) ([]model.Member, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error)
	GetByIDPublic(ctx context.Context, id uuid.UUID) (*model.PublicMember, error)
	GetByEmail(ctx context.Context, email string) (*model.Member, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error)
	Delete(ctx context.Context, id uuid.UUID) error
	HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error)
	DistinctSkills(ctx context.Context) ([]string, error)
}

// EmailSender sends emails via Resend.
type EmailSender interface {
	SendInvitation(ctx context.Context, toEmail, toName, frontendURL string)
}

type MemberService struct {
	repo        MemberRepo
	email       EmailSender
	frontendURL string
}

func NewMemberService(repo MemberRepo, email EmailSender, frontendURL string) *MemberService {
	return &MemberService{repo: repo, email: email, frontendURL: frontendURL}
}

func (s *MemberService) Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error) {
	// Validate
	details := map[string]string{}
	if req.Email == "" {
		details["email"] = "required"
	} else if _, err := mail.ParseAddress(req.Email); err != nil {
		details["email"] = "invalid email format"
	}
	if req.Name == "" {
		details["name"] = "required"
	}
	if len(details) > 0 {
		return nil, &ValidationError{Details: details}
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Strip @ from telegram handle
	if req.TelegramHandle != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*req.TelegramHandle), "@")
		if h == "" {
			req.TelegramHandle = nil
		} else {
			req.TelegramHandle = &h
		}
	}

	// Check for duplicate email
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if existing != nil {
		return nil, ErrDuplicateEmail
	}

	member, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("creating member: %w", err)
	}

	// Fire-and-forget invitation email
	if req.SendInvite && s.email != nil {
		go func() {
			s.email.SendInvitation(context.Background(), member.Email, member.Name, s.frontendURL)
		}()
	}

	return member, nil
}

func (s *MemberService) List(ctx context.Context, activeFilter string) ([]model.Member, error) {
	return s.repo.List(ctx, activeFilter)
}

func (s *MemberService) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotFound
	}
	return member, nil
}

func (s *MemberService) Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error) {
	// Validate name if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, &ValidationError{Details: map[string]string{"name": "cannot be empty"}}
		}
		req.Name = &name
	}

	// Strip @ from telegram handle
	if req.TelegramHandle != nil {
		h := strings.TrimPrefix(strings.TrimSpace(*req.TelegramHandle), "@")
		req.TelegramHandle = &h
	}

	member, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("updating member: %w", err)
	}
	if member == nil {
		return nil, ErrNotFound
	}
	return member, nil
}

func (s *MemberService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check member exists
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("checking member: %w", err)
	}
	if member == nil {
		return ErrNotFound
	}

	// Check for RSVPs
	hasRSVPs, err := s.repo.HasRSVPs(ctx, id)
	if err != nil {
		return fmt.Errorf("checking rsvps: %w", err)
	}
	if hasRSVPs {
		return ErrHasRSVPs
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting member: %w", err)
	}
	return nil
}

func (s *MemberService) ListDistinctSkills(ctx context.Context) ([]string, error) {
	return s.repo.DistinctSkills(ctx)
}

// ValidationError holds field-level validation details.
type ValidationError struct {
	Details map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

// NoopEmailSender is a no-op implementation used when email is not configured.
type NoopEmailSender struct{}

func (n *NoopEmailSender) SendInvitation(ctx context.Context, toEmail, toName, frontendURL string) {
	slog.Debug("invitation email skipped (noop)", "email", toEmail)
}
