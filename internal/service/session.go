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
	ErrSeriesNotFound     = errors.New("series not found")
	ErrSeriesInactive     = errors.New("series is already inactive")
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
	UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SpaceSession, error)
	ListFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error)
	UpdateBulkBySeriesID(ctx context.Context, seriesID uuid.UUID, req model.UpdateSessionRequest, imageURL *string) (int64, error)
	CancelFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error)
}

// SeriesRepo defines the data access interface for session series.
type SeriesRepo interface {
	Create(ctx context.Context, series model.SessionSeries) (*model.SessionSeries, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.SessionSeries, error)
	ListActive(ctx context.Context) ([]model.SessionSeries, error)
	Deactivate(ctx context.Context, id uuid.UUID) error
	GenerateSessions(ctx context.Context, series model.SessionSeries, fromDate, toDate time.Time) ([]model.SpaceSession, error)
	Update(ctx context.Context, id uuid.UUID, req model.UpdateSeriesRequest) (*model.SessionSeries, error)
	UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SessionSeries, error)
}

type SessionService struct {
	repo        SessionRepo
	seriesRepo  SeriesRepo
	notifier    Notifier
	emailSender NotificationEmailSender
	rsvpLister  RSVPMemberLister
}

func NewSessionService(repo SessionRepo, notifier Notifier) *SessionService {
	return &SessionService{repo: repo, notifier: notifier}
}

// SetSeriesRepo configures the session service to support recurring series.
func (s *SessionService) SetSeriesRepo(seriesRepo SeriesRepo) {
	s.seriesRepo = seriesRepo
}

func (s *SessionService) Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (any, error) {
	if err := s.validateCreate(req); err != nil {
		return nil, err
	}

	if req.RepeatForever {
		return s.createSeries(ctx, req, createdBy)
	}

	// Recurring sessions (finite)
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

// resolveRecurrenceStart returns the effective day-of-week and the first session date.
// If dayOfWeek is nil, it uses the weekday of baseDate. If dayOfWeek differs from baseDate,
// it advances to the next matching weekday.
func resolveRecurrenceStart(baseDate time.Time, dayOfWeek *int) (int, time.Time) {
	if dayOfWeek == nil {
		return int(baseDate.Weekday()), baseDate
	}
	targetDay := time.Weekday(*dayOfWeek)
	if baseDate.Weekday() == targetDay {
		return *dayOfWeek, baseDate
	}
	// Advance to the next matching weekday
	d := baseDate
	for d.Weekday() != targetDay {
		d = d.AddDate(0, 0, 1)
	}
	return *dayOfWeek, d
}

// effectiveInterval returns the recurrence interval in weeks, treating 0 as 1.
func effectiveInterval(everyNWeeks int) int {
	if everyNWeeks < 1 {
		return 1
	}
	return everyNWeeks
}

func (s *SessionService) createRecurring(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	baseDate, _ := time.Parse("2006-01-02", req.Date)
	dayOfWeek, startDate := resolveRecurrenceStart(baseDate, req.DayOfWeek)
	interval := effectiveInterval(req.EveryNWeeks)

	// Create a series record (inactive — finite, no auto-extension)
	series := model.SessionSeries{
		Title:       req.Title,
		Description: req.Description,
		DayOfWeek:   dayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		EveryNWeeks: interval,
		CreatedBy:   createdBy,
	}

	created, err := s.seriesRepo.Create(ctx, series)
	if err != nil {
		return nil, fmt.Errorf("creating session series for recurring: %w", err)
	}

	// Deactivate immediately — finite series don't auto-extend
	_ = s.seriesRepo.Deactivate(ctx, created.ID)

	var reqs []model.CreateSessionRequest
	for i := 0; i <= req.RepeatWeekly; i++ {
		sessionReq := model.CreateSessionRequest{
			Title:       req.Title,
			Description: req.Description,
			Date:        startDate.AddDate(0, 0, interval*7*i).Format("2006-01-02"),
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Location:    req.Location,
			SeriesID:    &created.ID,
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

func (s *SessionService) createSeries(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	baseDate, _ := time.Parse("2006-01-02", req.Date)
	dayOfWeek, startDate := resolveRecurrenceStart(baseDate, req.DayOfWeek)
	interval := effectiveInterval(req.EveryNWeeks)

	series := model.SessionSeries{
		Title:       req.Title,
		Description: req.Description,
		DayOfWeek:   dayOfWeek,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		EveryNWeeks: interval,
		CreatedBy:   createdBy,
	}

	created, err := s.seriesRepo.Create(ctx, series)
	if err != nil {
		return nil, fmt.Errorf("creating session series: %w", err)
	}

	// Generate sessions for the next 60+ days
	toDate := startDate.AddDate(0, 0, 67)
	sessions, err := s.seriesRepo.GenerateSessions(ctx, *created, startDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("generating series sessions: %w", err)
	}

	go s.notifier.SessionsCreatedRecurring(sessions)

	return sessions, nil
}

func (s *SessionService) List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error) {
	s.extendActiveSeries(ctx, to)
	return s.repo.List(ctx, from, to, status, memberID)
}

// extendActiveSeries lazily generates sessions for all active series up to the given date.
func (s *SessionService) extendActiveSeries(ctx context.Context, to string) {
	if s.seriesRepo == nil {
		return
	}

	activeSeries, err := s.seriesRepo.ListActive(ctx)
	if err != nil {
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	var toDate time.Time
	if to != "" {
		toDate, err = time.Parse("2006-01-02", to)
		if err != nil {
			toDate = today.AddDate(0, 0, 60)
		}
	} else {
		toDate = today.AddDate(0, 0, 60)
	}
	// Extend a week beyond the request to ensure coverage
	toDate = toDate.AddDate(0, 0, 7)

	for _, series := range activeSeries {
		s.seriesRepo.GenerateSessions(ctx, series, today, toDate)
	}
}

func (s *SessionService) CancelSeries(ctx context.Context, seriesID uuid.UUID) error {
	if s.seriesRepo == nil {
		return ErrSeriesNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return ErrSeriesNotFound
	}

	// Cancel all future sessions
	canceled, err := s.repo.CancelFutureBySeriesID(ctx, seriesID)
	if err != nil {
		return fmt.Errorf("canceling future sessions: %w", err)
	}

	// Deactivate series
	if series.IsActive {
		_ = s.seriesRepo.Deactivate(ctx, seriesID)
	}

	// Send notifications
	if len(canceled) > 0 {
		go s.notifier.SeriesCanceled(series, canceled)
		go s.sendSeriesCancelEmails(series, canceled)
	}

	return nil
}

func (s *SessionService) UpdateSeries(ctx context.Context, seriesID uuid.UUID, req model.UpdateSeriesRequest) (*model.SessionSeries, error) {
	if s.seriesRepo == nil {
		return nil, ErrSeriesNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return nil, ErrSeriesNotFound
	}

	// Update series template
	updated, err := s.seriesRepo.Update(ctx, seriesID, req)
	if err != nil {
		return nil, fmt.Errorf("updating series: %w", err)
	}

	// Propagate changes to future sessions
	sessionReq := model.UpdateSessionRequest{
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
	}
	_, err = s.repo.UpdateBulkBySeriesID(ctx, seriesID, sessionReq, nil)
	if err != nil {
		return nil, fmt.Errorf("propagating series update: %w", err)
	}

	// Send notifications
	affected, _ := s.repo.ListFutureBySeriesID(ctx, seriesID)
	if len(affected) > 0 {
		go s.notifier.SeriesUpdated(updated, affected)
		go s.sendSeriesUpdatedEmails(updated, affected)
	}

	return updated, nil
}

func (s *SessionService) UpdateSeriesImageURL(ctx context.Context, seriesID uuid.UUID, imageURL string) (*model.SessionSeries, error) {
	if s.seriesRepo == nil {
		return nil, ErrSeriesNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return nil, ErrSeriesNotFound
	}

	updated, err := s.seriesRepo.UpdateImageURL(ctx, seriesID, &imageURL)
	if err != nil {
		return nil, fmt.Errorf("updating series image: %w", err)
	}

	// Propagate to future sessions
	_, _ = s.repo.UpdateBulkBySeriesID(ctx, seriesID, model.UpdateSessionRequest{}, &imageURL)

	return updated, nil
}

func (s *SessionService) ClearSeriesImageURL(ctx context.Context, seriesID uuid.UUID) (*model.SessionSeries, error) {
	if s.seriesRepo == nil {
		return nil, ErrSeriesNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return nil, ErrSeriesNotFound
	}

	updated, err := s.seriesRepo.UpdateImageURL(ctx, seriesID, nil)
	if err != nil {
		return nil, fmt.Errorf("clearing series image: %w", err)
	}

	return updated, nil
}

// GetSeriesImageURL returns the current image URL for a series, or empty string if none.
func (s *SessionService) GetSeriesImageURL(ctx context.Context, seriesID uuid.UUID) (string, error) {
	if s.seriesRepo == nil {
		return "", ErrSeriesNotFound
	}

	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return "", fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return "", ErrSeriesNotFound
	}
	if series.ImageURL != nil {
		return *series.ImageURL, nil
	}
	return "", nil
}

func (s *SessionService) GetSeriesByID(ctx context.Context, seriesID uuid.UUID) (*model.SessionSeries, error) {
	if s.seriesRepo == nil {
		return nil, ErrSeriesNotFound
	}
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("getting series: %w", err)
	}
	if series == nil {
		return nil, ErrSeriesNotFound
	}
	return series, nil
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
	if req.RepeatForever && req.RepeatWeekly > 0 {
		details["repeat_weekly"] = "cannot use both repeat_weekly and repeat_forever"
	} else if !req.RepeatForever && (req.RepeatWeekly < 0 || req.RepeatWeekly > 12) {
		details["repeat_weekly"] = "must be between 0 and 12"
	}

	isRecurring := req.RepeatForever || req.RepeatWeekly > 0
	if req.DayOfWeek != nil {
		if !isRecurring {
			details["day_of_week"] = "only valid with repeat_weekly or repeat_forever"
		} else if *req.DayOfWeek < 0 || *req.DayOfWeek > 6 {
			details["day_of_week"] = "must be between 0 (Sunday) and 6 (Saturday)"
		}
	}
	if req.EveryNWeeks != 0 {
		if !isRecurring {
			details["every_n_weeks"] = "only valid with repeat_weekly or repeat_forever"
		} else if req.EveryNWeeks < 1 || req.EveryNWeeks > 4 {
			details["every_n_weeks"] = "must be between 1 and 4"
		}
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

func (s *SessionService) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL string) (*model.SpaceSession, error) {
	session, err := s.repo.GetByID(ctx, id, nil)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	updated, err := s.repo.UpdateImageURL(ctx, id, &imageURL)
	if err != nil {
		return nil, fmt.Errorf("updating image url: %w", err)
	}
	return updated, nil
}

func (s *SessionService) ClearImageURL(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	session, err := s.repo.GetByID(ctx, id, nil)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	updated, err := s.repo.UpdateImageURL(ctx, id, nil)
	if err != nil {
		return nil, fmt.Errorf("clearing image url: %w", err)
	}
	return updated, nil
}

// GetImageURL returns the current image URL for a session, or empty string if none.
func (s *SessionService) GetImageURL(ctx context.Context, id uuid.UUID) (string, error) {
	session, err := s.repo.GetByID(ctx, id, nil)
	if err != nil {
		return "", fmt.Errorf("getting session: %w", err)
	}
	if session == nil {
		return "", ErrSessionNotFound
	}
	if session.ImageURL != nil {
		return *session.ImageURL, nil
	}
	return "", nil
}
