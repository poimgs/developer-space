package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionCanceled    = errors.New("session is canceled")
	ErrAlreadyCanceled    = errors.New("session is already canceled")
	ErrCapacityBelowRSVPs = errors.New("cannot reduce capacity below current RSVP count")
)

// SessionRepo defines the data access interface for sessions.
type SessionRepo interface {
	Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (*model.SpaceSession, error)
	CreateBatch(ctx context.Context, sessions []model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error)
	List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error)
	GetByID(ctx context.Context, id uuid.UUID, memberID *uuid.UUID) (*model.SpaceSession, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateSessionRequest, newStatus *string) (*model.SpaceSession, error)
	Cancel(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error)
	GetRSVPCount(ctx context.Context, sessionID uuid.UUID) (int, error)
}

type SessionService struct {
	repo        SessionRepo
	notifier    Notifier
	emailSender NotificationEmailSender
	rsvpLister  RSVPMemberLister
}

func NewSessionService(repo SessionRepo, notifier Notifier) *SessionService {
	return &SessionService{repo: repo, notifier: notifier}
}

func (s *SessionService) Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (any, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	// Recurring sessions
	if req.RepeatWeekly > 0 {
		return s.createRecurring(ctx, req, createdBy)
	}

	session, err := s.repo.Create(ctx, req, createdBy)
	if err != nil {
		return nil, fmt.Errorf("creating session: %w", err)
	}

	go s.notifier.SessionCreated(session)

	return session, nil
}

func (s *SessionService) createRecurring(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	baseDate, _ := time.Parse("2006-01-02", req.Date)

	var reqs []model.CreateSessionRequest
	for i := 0; i <= req.RepeatWeekly; i++ {
		sessionReq := model.CreateSessionRequest{
			Title:       req.Title,
			Description: req.Description,
			Date:        baseDate.AddDate(0, 0, 7*i).Format("2006-01-02"),
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Capacity:    req.Capacity,
		}
		reqs = append(reqs, sessionReq)
	}

	sessions, err := s.repo.CreateBatch(ctx, reqs, createdBy)
	if err != nil {
		return nil, fmt.Errorf("creating recurring sessions: %w", err)
	}

	go s.notifier.SessionsCreatedRecurring(sessions)

	return sessions, nil
}

func (s *SessionService) List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error) {
	return s.repo.List(ctx, from, to, status, memberID)
}

func (s *SessionService) GetByID(ctx context.Context, id uuid.UUID, memberID *uuid.UUID) (*model.SpaceSession, error) {
	session, err := s.repo.GetByID(ctx, id, memberID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *SessionService) Update(ctx context.Context, id uuid.UUID, req model.UpdateSessionRequest) (*model.SpaceSession, error) {
	existing, err := s.repo.GetByID(ctx, id, nil)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	if existing == nil {
		return nil, ErrSessionNotFound
	}

	if existing.Status == "canceled" {
		return nil, ErrSessionCanceled
	}

	// Validate the update
	if err := s.validateUpdate(req, existing); err != nil {
		return nil, err
	}

	// Check capacity against current RSVP count
	if req.Capacity != nil {
		rsvpCount, err := s.repo.GetRSVPCount(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("checking rsvp count: %w", err)
		}
		if *req.Capacity < rsvpCount {
			return nil, ErrCapacityBelowRSVPs
		}
	}

	// Determine if status should change to "shifted"
	var newStatus *string
	dateTimeChanged := req.Date != nil || req.StartTime != nil || req.EndTime != nil
	if dateTimeChanged && existing.Status == "scheduled" {
		shifted := "shifted"
		newStatus = &shifted
	} else if dateTimeChanged && existing.Status == "shifted" {
		// stays shifted
		shifted := "shifted"
		newStatus = &shifted
	}

	session, err := s.repo.Update(ctx, id, req, newStatus)
	if err != nil {
		return nil, fmt.Errorf("updating session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Get rsvp_count for the response
	session.RSVPCount = existing.RSVPCount

	if dateTimeChanged {
		go s.notifier.SessionShifted(session)
		go s.sendShiftedEmails(existing, session)
	}

	return session, nil
}

func (s *SessionService) Cancel(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	existing, err := s.repo.GetByID(ctx, id, nil)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	if existing == nil {
		return nil, ErrSessionNotFound
	}

	if existing.Status == "canceled" {
		return nil, ErrAlreadyCanceled
	}

	session, err := s.repo.Cancel(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("canceling session: %w", err)
	}

	go s.notifier.SessionCanceled(session)
	go s.sendCancelEmails(session)

	return session, nil
}

func (s *SessionService) validateCreate(req model.CreateSessionRequest) error {
	details := map[string]string{}

	if req.Title == "" {
		details["title"] = "required"
	}
	if req.Date == "" {
		details["date"] = "required"
	} else {
		date, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			details["date"] = "invalid format, expected YYYY-MM-DD"
		} else {
			today := time.Now().Truncate(24 * time.Hour)
			if date.Before(today) {
				details["date"] = "must be today or later"
			}
		}
	}
	if req.StartTime == "" {
		details["start_time"] = "required"
	} else if _, err := time.Parse("15:04", req.StartTime); err != nil {
		details["start_time"] = "invalid format, expected HH:MM"
	}
	if req.EndTime == "" {
		details["end_time"] = "required"
	} else if _, err := time.Parse("15:04", req.EndTime); err != nil {
		details["end_time"] = "invalid format, expected HH:MM"
	}
	if req.StartTime != "" && req.EndTime != "" {
		if req.EndTime <= req.StartTime {
			details["end_time"] = "must be after start_time"
		}
	}
	if req.Capacity <= 0 {
		details["capacity"] = "must be greater than 0"
	}
	if req.RepeatWeekly < 0 || req.RepeatWeekly > 12 {
		details["repeat_weekly"] = "must be between 0 and 12"
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}
	return nil
}

func (s *SessionService) validateUpdate(req model.UpdateSessionRequest, existing *model.SpaceSession) error {
	details := map[string]string{}

	if req.Title != nil && *req.Title == "" {
		details["title"] = "cannot be empty"
	}

	// Resolve effective date/times for cross-field validation
	effectiveDate := existing.Date
	if req.Date != nil {
		effectiveDate = *req.Date
	}
	effectiveStartTime := existing.StartTime
	if req.StartTime != nil {
		effectiveStartTime = *req.StartTime
	}
	effectiveEndTime := existing.EndTime
	if req.EndTime != nil {
		effectiveEndTime = *req.EndTime
	}

	if req.Date != nil {
		date, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			details["date"] = "invalid format, expected YYYY-MM-DD"
		} else {
			today := time.Now().Truncate(24 * time.Hour)
			if date.Before(today) {
				details["date"] = "must be today or later"
			}
		}
	}
	if req.StartTime != nil {
		if _, err := time.Parse("15:04", *req.StartTime); err != nil {
			details["start_time"] = "invalid format, expected HH:MM"
		}
	}
	if req.EndTime != nil {
		if _, err := time.Parse("15:04", *req.EndTime); err != nil {
			details["end_time"] = "invalid format, expected HH:MM"
		}
	}

	// Cross-field: end must be after start
	if effectiveEndTime <= effectiveStartTime {
		if _, exists := details["end_time"]; !exists {
			if _, exists := details["start_time"]; !exists {
				details["end_time"] = "must be after start_time"
			}
		}
	}

	if req.Capacity != nil && *req.Capacity <= 0 {
		details["capacity"] = "must be greater than 0"
	}

	// Validate the effective date is not in the past
	if req.Date == nil {
		// Check existing date isn't past when changing times
		if req.StartTime != nil || req.EndTime != nil {
			date, err := time.Parse("2006-01-02", effectiveDate)
			if err == nil {
				today := time.Now().Truncate(24 * time.Hour)
				if date.Before(today) {
					details["date"] = "cannot edit a past session"
				}
			}
		}
	}

	if len(details) > 0 {
		return &ValidationError{Details: details}
	}
	return nil
}
