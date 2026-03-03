package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/service"
)

func setupProfileHandler() (*ProfileHandler, *service.AuthService, *mockMemberRepo) {
	tokenRepo := &mockTokenRepoForHandler{}
	memberRepo := newMockRepo()
	emailSender := &mockMagicLinkSenderForHandler{}
	authSvc := service.NewAuthService(tokenRepo, memberRepo, emailSender, "test-secret", "http://localhost:5173", false)
	h := NewProfileHandler(authSvc)
	return h, authSvc, memberRepo
}

func setupProfileRouter(h *ProfileHandler, authSvc *service.AuthService, memberRepo *mockMemberRepo) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc, memberRepo))
		r.Get("/api/profiles/{id}", h.GetPublicProfile)
	})
	return r
}

func TestProfileHandler_GetPublicProfile_200(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	// Create the authenticated caller
	callerID := uuid.New()
	memberRepo.members[callerID] = &model.Member{ID: callerID, Email: "caller@example.com", Name: "Caller", IsActive: true}

	// Create the target profile
	targetID := uuid.New()
	bio := "Hello world"
	skills := []string{"go", "react"}
	memberRepo.members[targetID] = &model.Member{
		ID:       targetID,
		Email:    "jane@example.com",
		Name:     "Jane Doe",
		IsActive: true,
		Bio:      &bio,
		Skills:   skills,
	}

	cookie, _ := authSvc.CreateSessionCookie(callerID)

	req := httptest.NewRequest("GET", "/api/profiles/"+targetID.String(), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.PublicMember `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data.ID != targetID {
		t.Errorf("expected ID %s, got %s", targetID, resp.Data.ID)
	}
	if resp.Data.Name != "Jane Doe" {
		t.Errorf("expected name 'Jane Doe', got %s", resp.Data.Name)
	}
	if resp.Data.Bio == nil || *resp.Data.Bio != "Hello world" {
		t.Errorf("expected bio 'Hello world', got %v", resp.Data.Bio)
	}
	if len(resp.Data.Skills) != 2 {
		t.Errorf("expected 2 skills, got %d", len(resp.Data.Skills))
	}
}

func TestProfileHandler_GetPublicProfile_404_NotFound(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	callerID := uuid.New()
	memberRepo.members[callerID] = &model.Member{ID: callerID, Email: "caller@example.com", Name: "Caller", IsActive: true}

	cookie, _ := authSvc.CreateSessionCookie(callerID)

	req := httptest.NewRequest("GET", "/api/profiles/"+uuid.New().String(), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProfileHandler_GetPublicProfile_404_InactiveMember(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	callerID := uuid.New()
	memberRepo.members[callerID] = &model.Member{ID: callerID, Email: "caller@example.com", Name: "Caller", IsActive: true}

	// Create inactive target
	targetID := uuid.New()
	memberRepo.members[targetID] = &model.Member{
		ID:       targetID,
		Email:    "inactive@example.com",
		Name:     "Inactive",
		IsActive: false,
	}

	cookie, _ := authSvc.CreateSessionCookie(callerID)

	req := httptest.NewRequest("GET", "/api/profiles/"+targetID.String(), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for inactive member, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProfileHandler_GetPublicProfile_400_InvalidID(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	callerID := uuid.New()
	memberRepo.members[callerID] = &model.Member{ID: callerID, Email: "caller@example.com", Name: "Caller", IsActive: true}

	cookie, _ := authSvc.CreateSessionCookie(callerID)

	req := httptest.NewRequest("GET", "/api/profiles/not-a-uuid", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProfileHandler_GetPublicProfile_401_Unauthenticated(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	req := httptest.NewRequest("GET", "/api/profiles/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProfileHandler_GetPublicProfile_NoPrivateFields(t *testing.T) {
	h, authSvc, memberRepo := setupProfileHandler()
	r := setupProfileRouter(h, authSvc, memberRepo)

	callerID := uuid.New()
	memberRepo.members[callerID] = &model.Member{ID: callerID, Email: "caller@example.com", Name: "Caller", IsActive: true}

	targetID := uuid.New()
	memberRepo.members[targetID] = &model.Member{
		ID:       targetID,
		Email:    "jane@example.com",
		Name:     "Jane Doe",
		IsAdmin:  true,
		IsActive: true,
	}

	cookie, _ := authSvc.CreateSessionCookie(callerID)

	req := httptest.NewRequest("GET", "/api/profiles/"+targetID.String(), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Unmarshal into raw map to verify no private fields leak
	var resp struct {
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// Should NOT contain email, is_admin, is_active, created_at, updated_at
	for _, field := range []string{"email", "is_admin", "is_active", "created_at", "updated_at"} {
		if _, exists := resp.Data[field]; exists {
			t.Errorf("public profile should NOT contain %s", field)
		}
	}

	// Should contain public fields
	for _, field := range []string{"id", "name"} {
		if _, exists := resp.Data[field]; !exists {
			t.Errorf("public profile should contain %s", field)
		}
	}
}
