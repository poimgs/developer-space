package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/service"
)

// mockMemberRepo implements service.MemberRepo for handler tests.
type mockMemberRepo struct {
	members    map[uuid.UUID]*model.Member
	byEmail    map[string]*model.Member
	rsvpCounts map[uuid.UUID]int
}

func newMockRepo() *mockMemberRepo {
	return &mockMemberRepo{
		members:    make(map[uuid.UUID]*model.Member),
		byEmail:    make(map[string]*model.Member),
		rsvpCounts: make(map[uuid.UUID]int),
	}
}

func (m *mockMemberRepo) Create(ctx context.Context, req model.CreateMemberRequest) (*model.Member, error) {
	member := &model.Member{
		ID:             uuid.New(),
		Email:          req.Email,
		Name:           req.Name,
		TelegramHandle: req.TelegramHandle,
		IsAdmin:        req.IsAdmin,
		IsActive:       true,
	}
	m.members[member.ID] = member
	m.byEmail[member.Email] = member
	return member, nil
}

func (m *mockMemberRepo) List(ctx context.Context, activeFilter string) ([]model.Member, error) {
	var result []model.Member
	for _, member := range m.members {
		result = append(result, *member)
	}
	if result == nil {
		result = []model.Member{}
	}
	return result, nil
}

func (m *mockMemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	return member, nil
}

func (m *mockMemberRepo) GetByIDPublic(ctx context.Context, id uuid.UUID) (*model.PublicMember, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	if !member.IsActive {
		return nil, nil
	}
	return &model.PublicMember{
		ID:              member.ID,
		Name:            member.Name,
		TelegramHandle:  member.TelegramHandle,
		Bio:             member.Bio,
		Skills:          member.Skills,
		LinkedinURL:     member.LinkedinURL,
		InstagramHandle: member.InstagramHandle,
		GithubUsername:  member.GithubUsername,
	}, nil
}

func (m *mockMemberRepo) GetByEmail(ctx context.Context, email string) (*model.Member, error) {
	member, ok := m.byEmail[email]
	if !ok {
		return nil, nil
	}
	return member, nil
}

func (m *mockMemberRepo) Update(ctx context.Context, id uuid.UUID, req model.UpdateMemberRequest) (*model.Member, error) {
	member, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	if req.Name != nil {
		member.Name = *req.Name
	}
	return member, nil
}

func (m *mockMemberRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.members, id)
	return nil
}

func (m *mockMemberRepo) HasRSVPs(ctx context.Context, memberID uuid.UUID) (bool, error) {
	return m.rsvpCounts[memberID] > 0, nil
}

type noopEmailSender struct{}

func (n *noopEmailSender) SendInvitation(ctx context.Context, toEmail, toName, frontendURL string) {}

func setupHandler() (*MemberHandler, *mockMemberRepo) {
	repo := newMockRepo()
	svc := service.NewMemberService(repo, &noopEmailSender{}, "http://localhost:5173")
	h := NewMemberHandler(svc)
	return h, repo
}

func setupRouter(h *MemberHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/members", h.Create)
	r.Get("/api/members", h.List)
	r.Get("/api/members/{id}", h.GetByID)
	r.Patch("/api/members/{id}", h.Update)
	r.Delete("/api/members/{id}", h.Delete)
	return r
}

func TestHandler_CreateMember_201(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	body := `{"email":"jane@example.com","name":"Jane Doe"}`
	req := httptest.NewRequest("POST", "/api/members", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]json.RawMessage
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["data"]; !ok {
		t.Error("expected response to have 'data' key")
	}
}

func TestHandler_CreateMember_422_MissingFields(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	body := `{}`
	req := httptest.NewRequest("POST", "/api/members", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]json.RawMessage
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["details"]; !ok {
		t.Error("expected response to have 'details' key")
	}
}

func TestHandler_CreateMember_409_DuplicateEmail(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	body := `{"email":"jane@example.com","name":"Jane Doe"}`

	// First create
	req := httptest.NewRequest("POST", "/api/members", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d", rec.Code)
	}

	// Duplicate
	req = httptest.NewRequest("POST", "/api/members", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ListMembers_200(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/members", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data []model.Member `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data == nil {
		t.Error("expected non-nil data array")
	}
}

func TestHandler_GetMember_404(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/members/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_GetMember_400_InvalidID(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	req := httptest.NewRequest("GET", "/api/members/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandler_UpdateMember_200(t *testing.T) {
	h, repo := setupHandler()
	r := setupRouter(h)

	// Seed a member directly in the mock
	id := uuid.New()
	repo.members[id] = &model.Member{ID: id, Email: "jane@example.com", Name: "Jane Doe", IsActive: true}
	repo.byEmail["jane@example.com"] = repo.members[id]

	body := `{"name":"Jane Smith"}`
	req := httptest.NewRequest("PATCH", "/api/members/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_UpdateMember_404(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	body := `{"name":"Jane Smith"}`
	req := httptest.NewRequest("PATCH", "/api/members/"+uuid.New().String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_DeleteMember_204(t *testing.T) {
	h, repo := setupHandler()
	r := setupRouter(h)

	id := uuid.New()
	repo.members[id] = &model.Member{ID: id, Email: "jane@example.com", Name: "Jane Doe", IsActive: true}

	req := httptest.NewRequest("DELETE", "/api/members/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_DeleteMember_409_HasRSVPs(t *testing.T) {
	h, repo := setupHandler()
	r := setupRouter(h)

	id := uuid.New()
	repo.members[id] = &model.Member{ID: id, Email: "jane@example.com", Name: "Jane Doe", IsActive: true}
	repo.rsvpCounts[id] = 2

	req := httptest.NewRequest("DELETE", "/api/members/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Error string `json:"error"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error != "Cannot delete member with existing RSVPs. Deactivate instead." {
		t.Errorf("unexpected error message: %s", resp.Error)
	}
}

func TestHandler_DeleteMember_404(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	req := httptest.NewRequest("DELETE", "/api/members/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandler_CreateMember_400_InvalidJSON(t *testing.T) {
	h, _ := setupHandler()
	r := setupRouter(h)

	body := `{invalid json}`
	req := httptest.NewRequest("POST", "/api/members", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
