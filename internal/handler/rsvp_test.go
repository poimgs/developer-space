package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/repository"
	"github.com/developer-space/api/internal/service"
)

// --- Mocks ---

type mockRSVPRepoForHandler struct {
	createResult *repository.RSVPTxResult
	createErr    error
	deleteResult *model.SpaceSession
	deleteErr    error
	listResult   []model.RSVPWithMember
	listErr      error
}

func (m *mockRSVPRepoForHandler) CreateAtomic(ctx context.Context, sessionID, memberID uuid.UUID) (*repository.RSVPTxResult, error) {
	return m.createResult, m.createErr
}

func (m *mockRSVPRepoForHandler) Delete(ctx context.Context, sessionID, memberID uuid.UUID) (*model.SpaceSession, error) {
	return m.deleteResult, m.deleteErr
}

func (m *mockRSVPRepoForHandler) ListBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPWithMember, error) {
	return m.listResult, m.listErr
}

type mockMemberGetterForHandler struct {
	member *model.Member
}

func (m *mockMemberGetterForHandler) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	return m.member, nil
}

type noopNotifierForHandler struct{}

func (n *noopNotifierForHandler) SessionCreated(session *model.SpaceSession)                          {}
func (n *noopNotifierForHandler) SessionsCreatedRecurring(sessions []model.SpaceSession)              {}
func (n *noopNotifierForHandler) SessionShifted(session *model.SpaceSession)                          {}
func (n *noopNotifierForHandler) SessionCanceled(session *model.SpaceSession)                         {}
func (n *noopNotifierForHandler) MemberRSVPed(session *model.SpaceSession, member *model.Member)      {}
func (n *noopNotifierForHandler) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {}
func (n *noopNotifierForHandler) SeriesUpdated(series *model.SessionSeries, affected []model.SpaceSession) {}
func (n *noopNotifierForHandler) SeriesCanceled(series *model.SessionSeries, canceled []model.SpaceSession) {}

// --- Helpers ---

func testHandlerMember() *model.Member {
	return &model.Member{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		IsAdmin:  false,
		IsActive: true,
	}
}

func testHandlerSession() *model.SpaceSession {
	return &model.SpaceSession{
		ID:        uuid.New(),
		Title:     "Coworking Session",
		Date:      time.Now().AddDate(0, 0, 7).Format("2006-01-02"),
		StartTime: "10:00",
		EndTime:   "18:00",
		Status:    "scheduled",
		CreatedBy: uuid.New(),
		RSVPCount: 3,
	}
}

func setupRSVPHandler(repo *mockRSVPRepoForHandler, member *model.Member) (*RSVPHandler, *chi.Mux) {
	memberGetter := &mockMemberGetterForHandler{member: member}
	svc := service.NewRSVPService(repo, memberGetter, &noopNotifierForHandler{})
	h := NewRSVPHandler(svc)

	r := chi.NewRouter()
	r.Route("/api/sessions/{id}", func(r chi.Router) {
		r.Post("/rsvp", func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.MemberKey, member)
			h.RSVP(w, req.WithContext(ctx))
		})
		r.Delete("/rsvp", func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.MemberKey, member)
			h.CancelRSVP(w, req.WithContext(ctx))
		})
		r.Get("/rsvps", func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), middleware.MemberKey, member)
			h.ListRSVPs(w, req.WithContext(ctx))
		})
	})

	return h, r
}

// --- RSVP Create tests ---

func TestRSVPHandler_RSVP_201(t *testing.T) {
	member := testHandlerMember()
	session := testHandlerSession()
	rsvp := &model.RSVP{
		ID:        uuid.New(),
		SessionID: session.ID,
		MemberID:  member.ID,
		CreatedAt: time.Now(),
	}

	repo := &mockRSVPRepoForHandler{
		createResult: &repository.RSVPTxResult{RSVP: rsvp, Session: session},
	}
	_, r := setupRSVPHandler(repo, member)

	req := httptest.NewRequest("POST", "/api/sessions/"+session.ID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]json.RawMessage
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["data"]; !ok {
		t.Fatal("expected 'data' key in response")
	}
}

func TestRSVPHandler_RSVP_404_SessionNotFound(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{createResult: nil, createErr: nil}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_RSVP_409_Duplicate(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{createErr: repository.ErrRSVPDuplicate}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_RSVP_422_Canceled(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{createErr: repository.ErrRSVPSessionCanceled}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_RSVP_422_Past(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{createErr: repository.ErrRSVPSessionPast}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_RSVP_400_InvalidID(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{}
	_, r := setupRSVPHandler(repo, member)

	req := httptest.NewRequest("POST", "/api/sessions/not-a-uuid/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Cancel RSVP tests ---

func TestRSVPHandler_CancelRSVP_204(t *testing.T) {
	member := testHandlerMember()
	session := testHandlerSession()
	repo := &mockRSVPRepoForHandler{deleteResult: session}
	_, r := setupRSVPHandler(repo, member)

	req := httptest.NewRequest("DELETE", "/api/sessions/"+session.ID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_CancelRSVP_404_SessionNotFound(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{deleteResult: nil, deleteErr: nil}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_CancelRSVP_404_NotRSVPed(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{deleteErr: repository.ErrRSVPNotFound}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_CancelRSVP_422_PastSession(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{deleteErr: repository.ErrRSVPSessionPast}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("DELETE", "/api/sessions/"+sessionID.String()+"/rsvp", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- List RSVPs tests ---

func TestRSVPHandler_ListRSVPs_200(t *testing.T) {
	member := testHandlerMember()
	sessionID := uuid.New()
	tg := "@user1"
	rsvps := []model.RSVPWithMember{
		{
			ID:        uuid.New(),
			SessionID: sessionID,
			Member:    model.RSVPMember{ID: uuid.New(), Name: "User One", TelegramHandle: &tg},
			CreatedAt: time.Now(),
		},
	}
	repo := &mockRSVPRepoForHandler{listResult: rsvps}
	_, r := setupRSVPHandler(repo, member)

	req := httptest.NewRequest("GET", "/api/sessions/"+sessionID.String()+"/rsvps", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]json.RawMessage
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["data"]; !ok {
		t.Fatal("expected 'data' key in response")
	}

	var data []model.RSVPWithMember
	json.Unmarshal(resp["data"], &data)
	if len(data) != 1 {
		t.Fatalf("expected 1 rsvp, got %d", len(data))
	}
}

func TestRSVPHandler_ListRSVPs_404_SessionNotFound(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{listResult: nil}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("GET", "/api/sessions/"+sessionID.String()+"/rsvps", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRSVPHandler_ListRSVPs_200_Empty(t *testing.T) {
	member := testHandlerMember()
	repo := &mockRSVPRepoForHandler{listResult: []model.RSVPWithMember{}}
	_, r := setupRSVPHandler(repo, member)

	sessionID := uuid.New()
	req := httptest.NewRequest("GET", "/api/sessions/"+sessionID.String()+"/rsvps", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
