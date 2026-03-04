package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

// --- Mock implementations ---

type mockSessionRepo struct {
	sessions    map[uuid.UUID]*model.SpaceSession
	rsvpCounts  map[uuid.UUID]int
	createErr   error
	batchCalled bool
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:   make(map[uuid.UUID]*model.SpaceSession),
		rsvpCounts: make(map[uuid.UUID]int),
	}
}

func (m *mockSessionRepo) Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (*model.SpaceSession, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	s := &model.SpaceSession{
		ID:        uuid.New(),
		Title:     req.Title,
		Description: req.Description,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Capacity:  req.Capacity,
		Status:    "scheduled",
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sessions[s.ID] = s
	return s, nil
}

func (m *mockSessionRepo) CreateBatch(ctx context.Context, sessions []model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	m.batchCalled = true
	var result []model.SpaceSession
	for _, req := range sessions {
		s := model.SpaceSession{
			ID:        uuid.New(),
			Title:     req.Title,
			Description: req.Description,
			Date:      req.Date,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			Capacity:  req.Capacity,
			Status:    "scheduled",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		m.sessions[s.ID] = &s
		result = append(result, s)
	}
	return result, nil
}

func (m *mockSessionRepo) List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error) {
	var result []model.SpaceSession
	for _, s := range m.sessions {
		result = append(result, *s)
	}
	if result == nil {
		result = []model.SpaceSession{}
	}
	return result, nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id uuid.UUID, memberID *uuid.UUID) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockSessionRepo) Update(ctx context.Context, id uuid.UUID, req model.UpdateSessionRequest, newStatus *string) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	if req.Title != nil {
		s.Title = *req.Title
	}
	if req.Description != nil {
		s.Description = req.Description
	}
	if req.Date != nil {
		s.Date = *req.Date
	}
	if req.StartTime != nil {
		s.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		s.EndTime = *req.EndTime
	}
	if req.Capacity != nil {
		s.Capacity = *req.Capacity
	}
	if newStatus != nil {
		s.Status = *newStatus
	}
	s.UpdatedAt = time.Now()
	return s, nil
}

func (m *mockSessionRepo) Cancel(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	s.Status = "canceled"
	s.UpdatedAt = time.Now()
	return s, nil
}

func (m *mockSessionRepo) GetRSVPCount(ctx context.Context, sessionID uuid.UUID) (int, error) {
	return m.rsvpCounts[sessionID], nil
}

func (m *mockSessionRepo) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	s.ImageURL = imageURL
	s.UpdatedAt = time.Now()
	return s, nil
}

// addSession adds a session to the mock repo and returns it.
func (m *mockSessionRepo) addSession(title, date, startTime, endTime, status string, capacity int) *model.SpaceSession {
	s := &model.SpaceSession{
		ID:        uuid.New(),
		Title:     title,
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Capacity:  capacity,
		Status:    status,
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sessions[s.ID] = s
	return s
}

type mockNotifier struct {
	createdCalls          int
	createdRecurCalls     int
	shiftedCalls          int
	canceledCalls         int
	lastRecurringSessions []model.SpaceSession
}

func (n *mockNotifier) SessionCreated(session *model.SpaceSession)                          { n.createdCalls++ }
func (n *mockNotifier) SessionsCreatedRecurring(sessions []model.SpaceSession)              { n.createdRecurCalls++; n.lastRecurringSessions = sessions }
func (n *mockNotifier) SessionShifted(session *model.SpaceSession)                          { n.shiftedCalls++ }
func (n *mockNotifier) SessionCanceled(session *model.SpaceSession)                         { n.canceledCalls++ }
func (n *mockNotifier) MemberRSVPed(session *model.SpaceSession, member *model.Member)      {}
func (n *mockNotifier) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {}

// --- Helper ---

func futureDate() string {
	return time.Now().AddDate(0, 0, 7).Format("2006-01-02")
}

func ptrStr(s string) *string { return &s }
func ptrInt(i int) *int       { return &i }

// --- Tests ---

func TestCreateSession_Valid(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	result, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Friday Session",
		Date:      futureDate(),
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  8,
	}, uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session, ok := result.(*model.SpaceSession)
	if !ok {
		t.Fatal("expected *SpaceSession result")
	}
	if session.Title != "Friday Session" {
		t.Errorf("expected title 'Friday Session', got %q", session.Title)
	}
	if session.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got %q", session.Status)
	}

	// Give goroutine time to fire
	time.Sleep(10 * time.Millisecond)
	if notifier.createdCalls != 1 {
		t.Errorf("expected 1 SessionCreated notification, got %d", notifier.createdCalls)
	}
}

func TestCreateSession_MissingTitle(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Date:      futureDate(),
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  8,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["title"] != "required" {
		t.Errorf("expected title required, got %q", ve.Details["title"])
	}
}

func TestCreateSession_InvalidDate(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Test",
		Date:      "not-a-date",
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  8,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if _, ok := ve.Details["date"]; !ok {
		t.Error("expected date validation error")
	}
}

func TestCreateSession_PastDate(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Test",
		Date:      "2020-01-01",
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  8,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["date"] != "must be today or later" {
		t.Errorf("expected 'must be today or later', got %q", ve.Details["date"])
	}
}

func TestCreateSession_EndBeforeStart(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Test",
		Date:      futureDate(),
		StartTime: "18:00",
		EndTime:   "14:00",
		Capacity:  8,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["end_time"] != "must be after start_time" {
		t.Errorf("expected end_time error, got %q", ve.Details["end_time"])
	}
}

func TestCreateSession_ZeroCapacity(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Test",
		Date:      futureDate(),
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  0,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["capacity"] != "must be greater than 0" {
		t.Errorf("expected capacity error, got %q", ve.Details["capacity"])
	}
}

func TestCreateSession_RepeatWeekly13(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Test",
		Date:         futureDate(),
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 13,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["repeat_weekly"] != "must be between 0 and 12" {
		t.Errorf("expected repeat_weekly error, got %q", ve.Details["repeat_weekly"])
	}
}

func TestCreateSession_Recurring(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	baseDate := futureDate()
	result, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Weekly Session",
		Date:         baseDate,
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 3,
	}, uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, ok := result.([]model.SpaceSession)
	if !ok {
		t.Fatal("expected []SpaceSession for recurring")
	}
	if len(sessions) != 4 {
		t.Fatalf("expected 4 sessions (N+1), got %d", len(sessions))
	}

	// Verify dates advance by 7 days each
	base, _ := time.Parse("2006-01-02", baseDate)
	for i, s := range sessions {
		expected := base.AddDate(0, 0, 7*i).Format("2006-01-02")
		if s.Date != expected {
			t.Errorf("session[%d]: expected date %s, got %s", i, expected, s.Date)
		}
		if s.Title != "Weekly Session" {
			t.Errorf("session[%d]: expected title 'Weekly Session', got %q", i, s.Title)
		}
	}

	if !repo.batchCalled {
		t.Error("expected CreateBatch to be called")
	}

	time.Sleep(10 * time.Millisecond)
	if notifier.createdRecurCalls != 1 {
		t.Errorf("expected 1 recurring notification, got %d", notifier.createdRecurCalls)
	}
}

func TestGetSessionByID_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.GetByID(context.Background(), uuid.New(), nil)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestGetSessionByID_Found(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Test", futureDate(), "14:00", "18:00", "scheduled", 8)

	session, err := svc.GetByID(context.Background(), existing.ID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", session.Title)
	}
}

func TestUpdateSession_ChangeTitle(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	existing := repo.addSession("Old Title", futureDate(), "14:00", "18:00", "scheduled", 8)

	session, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Title: ptrStr("New Title"),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Title != "New Title" {
		t.Errorf("expected 'New Title', got %q", session.Title)
	}
	if session.Status != "scheduled" {
		t.Errorf("expected status 'scheduled' (no date/time change), got %q", session.Status)
	}

	time.Sleep(10 * time.Millisecond)
	if notifier.shiftedCalls != 0 {
		t.Error("should NOT notify shifted for title-only change")
	}
}

func TestUpdateSession_ChangeDate_BecomesShifted(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	newDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	session, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Date: ptrStr(newDate),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "shifted" {
		t.Errorf("expected status 'shifted', got %q", session.Status)
	}

	time.Sleep(10 * time.Millisecond)
	if notifier.shiftedCalls != 1 {
		t.Errorf("expected 1 shifted notification, got %d", notifier.shiftedCalls)
	}
}

func TestUpdateSession_ChangeStartTime_BecomesShifted(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	session, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		StartTime: ptrStr("15:00"),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "shifted" {
		t.Errorf("expected status 'shifted', got %q", session.Status)
	}
}

func TestUpdateSession_AlreadyShifted_StaysShifted(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	newDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "shifted", 8)

	session, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Date: ptrStr(newDate),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "shifted" {
		t.Errorf("expected status 'shifted', got %q", session.Status)
	}
}

func TestUpdateSession_CanceledSession_Rejected(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "canceled", 8)

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Title: ptrStr("New Title"),
	})

	if !errors.Is(err, ErrSessionCanceled) {
		t.Errorf("expected ErrSessionCanceled, got %v", err)
	}
}

func TestUpdateSession_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Update(context.Background(), uuid.New(), model.UpdateSessionRequest{
		Title: ptrStr("New"),
	})

	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestUpdateSession_CapacityBelowRSVPs(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)
	repo.rsvpCounts[existing.ID] = 5

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Capacity: ptrInt(3),
	})

	if !errors.Is(err, ErrCapacityBelowRSVPs) {
		t.Errorf("expected ErrCapacityBelowRSVPs, got %v", err)
	}
}

func TestUpdateSession_CapacityAboveRSVPs_OK(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)
	repo.rsvpCounts[existing.ID] = 5

	session, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Capacity: ptrInt(6),
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Capacity != 6 {
		t.Errorf("expected capacity 6, got %d", session.Capacity)
	}
}

func TestUpdateSession_EmptyTitle_Rejected(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Title: ptrStr(""),
	})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["title"] != "cannot be empty" {
		t.Errorf("expected title error, got %q", ve.Details["title"])
	}
}

func TestCancelSession_Scheduled(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	session, err := svc.Cancel(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "canceled" {
		t.Errorf("expected status 'canceled', got %q", session.Status)
	}

	time.Sleep(10 * time.Millisecond)
	if notifier.canceledCalls != 1 {
		t.Errorf("expected 1 cancel notification, got %d", notifier.canceledCalls)
	}
}

func TestCancelSession_Shifted(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "shifted", 8)

	session, err := svc.Cancel(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "canceled" {
		t.Errorf("expected status 'canceled', got %q", session.Status)
	}
}

func TestCancelSession_AlreadyCanceled(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "canceled", 8)

	_, err := svc.Cancel(context.Background(), existing.ID)
	if !errors.Is(err, ErrAlreadyCanceled) {
		t.Errorf("expected ErrAlreadyCanceled, got %v", err)
	}
}

func TestCancelSession_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Cancel(context.Background(), uuid.New())
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestCreateSession_MultipleValidationErrors(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		// All fields missing/invalid
		Title:        "",
		Date:         "",
		StartTime:    "",
		EndTime:      "",
		Capacity:     0,
		RepeatWeekly: -1,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	// Should have errors for all fields
	expectedFields := []string{"title", "date", "start_time", "end_time", "capacity", "repeat_weekly"}
	for _, field := range expectedFields {
		if _, ok := ve.Details[field]; !ok {
			t.Errorf("expected validation error for %q", field)
		}
	}
}

func TestUpdateSession_InvalidDateFormat(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Date: ptrStr("bad-date"),
	})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if _, ok := ve.Details["date"]; !ok {
		t.Error("expected date validation error")
	}
}

func TestUpdateSession_EndTimeBeforeStartTime(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled", 8)

	// Change end_time to be before existing start_time
	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		EndTime: ptrStr("12:00"),
	})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["end_time"] != "must be after start_time" {
		t.Errorf("expected end_time error, got %q", ve.Details["end_time"])
	}
}

func TestCreateSession_RecurringEvery2Weeks(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	baseDate := futureDate()
	result, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Biweekly Session",
		Date:         baseDate,
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 3,
		EveryNWeeks:  2,
	}, uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, ok := result.([]model.SpaceSession)
	if !ok {
		t.Fatal("expected []SpaceSession for recurring")
	}
	if len(sessions) != 4 {
		t.Fatalf("expected 4 sessions (N+1), got %d", len(sessions))
	}

	// Verify dates advance by 14 days each
	base, _ := time.Parse("2006-01-02", baseDate)
	for i, s := range sessions {
		expected := base.AddDate(0, 0, 14*i).Format("2006-01-02")
		if s.Date != expected {
			t.Errorf("session[%d]: expected date %s, got %s", i, expected, s.Date)
		}
	}
}

func TestCreateSession_RecurringWithDayOfWeek(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)

	// Pick a date and choose a different weekday
	baseDate := futureDate()
	base, _ := time.Parse("2006-01-02", baseDate)
	// Target: next weekday after the base date's weekday (wrapping around)
	targetDay := (int(base.Weekday()) + 2) % 7

	result, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Shifted Day Session",
		Date:         baseDate,
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 2,
		DayOfWeek:    &targetDay,
	}, uuid.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sessions, ok := result.([]model.SpaceSession)
	if !ok {
		t.Fatal("expected []SpaceSession for recurring")
	}
	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions (N+1), got %d", len(sessions))
	}

	// Verify all sessions land on the target weekday
	for i, s := range sessions {
		d, _ := time.Parse("2006-01-02", s.Date)
		if int(d.Weekday()) != targetDay {
			t.Errorf("session[%d]: expected weekday %d, got %d (date %s)", i, targetDay, int(d.Weekday()), s.Date)
		}
	}

	// Verify first session is on or after baseDate
	first, _ := time.Parse("2006-01-02", sessions[0].Date)
	if first.Before(base) {
		t.Errorf("first session %s is before base date %s", sessions[0].Date, baseDate)
	}
}

func TestCreateSession_InvalidDayOfWeek(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	invalidDay := 7
	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Test",
		Date:         futureDate(),
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 2,
		DayOfWeek:    &invalidDay,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if _, ok := ve.Details["day_of_week"]; !ok {
		t.Error("expected day_of_week validation error")
	}
}

func TestCreateSession_DayOfWeekWithoutRepeat(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	day := 3
	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:     "Test",
		Date:      futureDate(),
		StartTime: "14:00",
		EndTime:   "18:00",
		Capacity:  8,
		DayOfWeek: &day,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["day_of_week"] != "only valid with repeat_weekly or repeat_forever" {
		t.Errorf("expected day_of_week error, got %q", ve.Details["day_of_week"])
	}
}

func TestCreateSession_InvalidEveryNWeeks(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:        "Test",
		Date:         futureDate(),
		StartTime:    "14:00",
		EndTime:      "18:00",
		Capacity:     8,
		RepeatWeekly: 2,
		EveryNWeeks:  5,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["every_n_weeks"] != "must be between 1 and 4" {
		t.Errorf("expected every_n_weeks error, got %q", ve.Details["every_n_weeks"])
	}
}

func TestCreateSession_EveryNWeeksWithoutRepeat(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	_, err := svc.Create(context.Background(), model.CreateSessionRequest{
		Title:       "Test",
		Date:        futureDate(),
		StartTime:   "14:00",
		EndTime:     "18:00",
		Capacity:    8,
		EveryNWeeks: 2,
	}, uuid.New())

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if ve.Details["every_n_weeks"] != "only valid with repeat_weekly or repeat_forever" {
		t.Errorf("expected every_n_weeks error, got %q", ve.Details["every_n_weeks"])
	}
}

func TestList_ReturnsEmptySlice(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	sessions, err := svc.List(context.Background(), "", "", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions == nil {
		t.Error("expected non-nil empty slice")
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

