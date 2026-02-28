package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/repository"
)

// --- Mocks ---

type mockRSVPRepo struct {
	createResult *repository.RSVPTxResult
	createErr    error
	deleteResult *model.SpaceSession
	deleteErr    error
	listResult   []model.RSVPWithMember
	listErr      error
}

func (m *mockRSVPRepo) CreateAtomic(ctx context.Context, sessionID, memberID uuid.UUID) (*repository.RSVPTxResult, error) {
	return m.createResult, m.createErr
}

func (m *mockRSVPRepo) Delete(ctx context.Context, sessionID, memberID uuid.UUID) (*model.SpaceSession, error) {
	return m.deleteResult, m.deleteErr
}

func (m *mockRSVPRepo) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPWithMember, error) {
	return m.listResult, m.listErr
}

type mockMemberGetter struct {
	member *model.Member
}

func (m *mockMemberGetter) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	return m.member, nil
}

type mockRSVPNotifier struct {
	rsvpedCalls       int
	canceledRSVPCalls int
}

func (n *mockRSVPNotifier) SessionCreated(session *model.SpaceSession)                     {}
func (n *mockRSVPNotifier) SessionsCreatedRecurring(sessions []model.SpaceSession)         {}
func (n *mockRSVPNotifier) SessionShifted(session *model.SpaceSession)                     {}
func (n *mockRSVPNotifier) SessionCanceled(session *model.SpaceSession)                    {}
func (n *mockRSVPNotifier) MemberRSVPed(session *model.SpaceSession, member *model.Member) { n.rsvpedCalls++ }
func (n *mockRSVPNotifier) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {
	n.canceledRSVPCalls++
}

// --- Helpers ---

func testMember() *model.Member {
	return &model.Member{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		IsActive: true,
	}
}

func testSession() *model.SpaceSession {
	return &model.SpaceSession{
		ID:        uuid.New(),
		Title:     "Coworking Session",
		Date:      time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		StartTime: "10:00",
		EndTime:   "18:00",
		Capacity:  8,
		Status:    "scheduled",
		CreatedBy: uuid.New(),
		RSVPCount: 3,
	}
}

func testRSVP(sessionID, memberID uuid.UUID) *model.RSVP {
	return &model.RSVP{
		ID:        uuid.New(),
		SessionID: sessionID,
		MemberID:  memberID,
		CreatedAt: time.Now(),
	}
}

// --- Create tests ---

func TestRSVPService_Create_Success(t *testing.T) {
	session := testSession()
	member := testMember()
	rsvp := testRSVP(session.ID, member.ID)

	repo := &mockRSVPRepo{
		createResult: &repository.RSVPTxResult{RSVP: rsvp, Session: session},
	}
	notifier := &mockRSVPNotifier{}
	svc := NewRSVPService(repo, &mockMemberGetter{member: member}, notifier)

	result, err := svc.Create(context.Background(), session.ID, member.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != rsvp.ID {
		t.Errorf("expected rsvp id %s, got %s", rsvp.ID, result.ID)
	}
	// Notification fires in goroutine — just verify no panic
	time.Sleep(10 * time.Millisecond)
	if notifier.rsvpedCalls != 1 {
		t.Errorf("expected 1 notifier call, got %d", notifier.rsvpedCalls)
	}
}

func TestRSVPService_Create_SessionNotFound(t *testing.T) {
	repo := &mockRSVPRepo{createResult: nil, createErr: nil}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionNotFound) {
		t.Fatalf("expected ErrRSVPSessionNotFound, got %v", err)
	}
}

func TestRSVPService_Create_SessionCanceled(t *testing.T) {
	repo := &mockRSVPRepo{createErr: repository.ErrRSVPSessionCanceled}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionCanceled) {
		t.Fatalf("expected ErrRSVPSessionCanceled, got %v", err)
	}
}

func TestRSVPService_Create_SessionPast(t *testing.T) {
	repo := &mockRSVPRepo{createErr: repository.ErrRSVPSessionPast}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionPast) {
		t.Fatalf("expected ErrRSVPSessionPast, got %v", err)
	}
}

func TestRSVPService_Create_SessionFull(t *testing.T) {
	repo := &mockRSVPRepo{createErr: repository.ErrRSVPSessionFull}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionFull) {
		t.Fatalf("expected ErrRSVPSessionFull, got %v", err)
	}
}

func TestRSVPService_Create_Duplicate(t *testing.T) {
	repo := &mockRSVPRepo{createErr: repository.ErrRSVPDuplicate}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.Create(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPDuplicate) {
		t.Fatalf("expected ErrRSVPDuplicate, got %v", err)
	}
}

// --- Cancel tests ---

func TestRSVPService_Cancel_Success(t *testing.T) {
	session := testSession()
	member := testMember()

	repo := &mockRSVPRepo{deleteResult: session}
	notifier := &mockRSVPNotifier{}
	svc := NewRSVPService(repo, &mockMemberGetter{member: member}, notifier)

	err := svc.Cancel(context.Background(), session.ID, member.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	if notifier.canceledRSVPCalls != 1 {
		t.Errorf("expected 1 notifier call, got %d", notifier.canceledRSVPCalls)
	}
}

func TestRSVPService_Cancel_SessionNotFound(t *testing.T) {
	repo := &mockRSVPRepo{deleteResult: nil, deleteErr: nil}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	err := svc.Cancel(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionNotFound) {
		t.Fatalf("expected ErrRSVPSessionNotFound, got %v", err)
	}
}

func TestRSVPService_Cancel_RSVPNotFound(t *testing.T) {
	repo := &mockRSVPRepo{deleteErr: repository.ErrRSVPNotFound}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	err := svc.Cancel(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPNotFound) {
		t.Fatalf("expected ErrRSVPNotFound, got %v", err)
	}
}

func TestRSVPService_Cancel_PastSession(t *testing.T) {
	repo := &mockRSVPRepo{deleteErr: repository.ErrRSVPSessionPast}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	err := svc.Cancel(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionPast) {
		t.Fatalf("expected ErrRSVPSessionPast, got %v", err)
	}
}

// --- ListBySession tests ---

func TestRSVPService_ListBySession_Success(t *testing.T) {
	sessionID := uuid.New()
	memberID := uuid.New()
	tg := "@testuser"
	rsvps := []model.RSVPWithMember{
		{
			ID:        uuid.New(),
			SessionID: sessionID,
			Member:    model.RSVPMember{ID: memberID, Name: "Test User", TelegramHandle: &tg},
			CreatedAt: time.Now(),
		},
	}

	repo := &mockRSVPRepo{listResult: rsvps}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	result, err := svc.ListBySession(context.Background(), sessionID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 rsvp, got %d", len(result))
	}
	if result[0].Member.Name != "Test User" {
		t.Errorf("expected member name 'Test User', got '%s'", result[0].Member.Name)
	}
}

func TestRSVPService_ListBySession_EmptyList(t *testing.T) {
	repo := &mockRSVPRepo{listResult: []model.RSVPWithMember{}}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	result, err := svc.ListBySession(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 rsvps, got %d", len(result))
	}
}

func TestRSVPService_ListBySession_SessionNotFound(t *testing.T) {
	repo := &mockRSVPRepo{listResult: nil}
	svc := NewRSVPService(repo, &mockMemberGetter{}, &mockRSVPNotifier{})

	_, err := svc.ListBySession(context.Background(), uuid.New())
	if !errors.Is(err, ErrRSVPSessionNotFound) {
		t.Fatalf("expected ErrRSVPSessionNotFound, got %v", err)
	}
}
