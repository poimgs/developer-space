package handler

import (
	"bytes"
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
	"github.com/developer-space/api/internal/service"
)

// --- Mock implementations for auth handler tests ---

type mockTokenRepoForHandler struct {
	tokens      []*model.MagicToken
	recentCount int
	createCalled bool
}

func (m *mockTokenRepoForHandler) Create(ctx context.Context, memberID uuid.UUID, tokenHash string, expiresAt time.Time) (*model.MagicToken, error) {
	m.createCalled = true
	t := &model.MagicToken{ID: uuid.New(), MemberID: memberID, TokenHash: tokenHash, ExpiresAt: expiresAt}
	m.tokens = append(m.tokens, t)
	return t, nil
}

func (m *mockTokenRepoForHandler) FindValidByHash(ctx context.Context, tokenHash string) (*model.MagicToken, error) {
	for _, t := range m.tokens {
		if t.TokenHash == tokenHash && t.UsedAt == nil && t.ExpiresAt.After(time.Now()) {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTokenRepoForHandler) MarkUsed(ctx context.Context, id uuid.UUID) error {
	for _, t := range m.tokens {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
		}
	}
	return nil
}

func (m *mockTokenRepoForHandler) CountRecentByEmail(ctx context.Context, email string) (int, error) {
	return m.recentCount, nil
}

func (m *mockTokenRepoForHandler) CleanExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

type mockMagicLinkSenderForHandler struct{}

func (m *mockMagicLinkSenderForHandler) SendMagicLink(ctx context.Context, toEmail, toName, link string) error {
	return nil
}

func setupAuthHandler() (*AuthHandler, *service.AuthService, *mockMemberRepo, *mockTokenRepoForHandler) {
	tokenRepo := &mockTokenRepoForHandler{}
	memberRepo := newMockRepo()
	emailSender := &mockMagicLinkSenderForHandler{}
	authSvc := service.NewAuthService(tokenRepo, memberRepo, emailSender, "test-secret", "http://localhost:5173", false)
	h := NewAuthHandler(authSvc)
	return h, authSvc, memberRepo, tokenRepo
}

// mockMemberRepo already implements service.MemberRepo from member_test.go

func setupAuthRouter(h *AuthHandler, authSvc *service.AuthService, memberRepo *mockMemberRepo) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/auth/magic-link", h.RequestMagicLink)
	r.Get("/api/auth/verify", h.Verify)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc, memberRepo))
		r.Get("/api/auth/me", h.Me)
		r.Post("/api/auth/logout", h.Logout)
		r.Patch("/api/auth/profile", h.UpdateProfile)
	})
	return r
}

func TestAuthHandler_RequestMagicLink_200(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	// Add a member
	id := uuid.New()
	memberRepo.members[id] = &model.Member{ID: id, Email: "jane@example.com", Name: "Jane", IsActive: true}
	memberRepo.byEmail["jane@example.com"] = memberRepo.members[id]

	body := `{"email":"jane@example.com"}`
	req := httptest.NewRequest("POST", "/api/auth/magic-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			Message string `json:"message"`
		} `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Message == "" {
		t.Error("expected message in response")
	}
}

func TestAuthHandler_RequestMagicLink_200_UnknownEmail(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	body := `{"email":"unknown@example.com"}`
	req := httptest.NewRequest("POST", "/api/auth/magic-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Should still return 200 (no enumeration)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAuthHandler_RequestMagicLink_400_BadJSON(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	body := `{bad json}`
	req := httptest.NewRequest("POST", "/api/auth/magic-link", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestAuthHandler_Verify_MissingToken(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	req := httptest.NewRequest("GET", "/api/auth/verify", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandler_Verify_InvalidToken(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	req := httptest.NewRequest("GET", "/api/auth/verify?token=invalid-token", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandler_Me_Unauthenticated(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Me_Authenticated(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	memberID := uuid.New()
	memberRepo.members[memberID] = &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true}

	// Create a valid session cookie
	cookie, err := authSvc.CreateSessionCookie(memberID)
	if err != nil {
		t.Fatalf("failed to create cookie: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/auth/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.Member `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, resp.Data.ID)
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	memberID := uuid.New()
	memberRepo.members[memberID] = &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true}

	cookie, _ := authSvc.CreateSessionCookie(memberID)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Check that a clearing cookie was set
	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session" && c.MaxAge < 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected session cookie to be cleared")
	}
}

func TestAuthHandler_UpdateProfile_200(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	memberID := uuid.New()
	memberRepo.members[memberID] = &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane Doe", IsActive: true}

	cookie, _ := authSvc.CreateSessionCookie(memberID)

	body := `{"name":"Jane Smith","telegram_handle":"@janesmith"}`
	req := httptest.NewRequest("PATCH", "/api/auth/profile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.Member `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Name != "Jane Smith" {
		t.Errorf("expected name 'Jane Smith', got %s", resp.Data.Name)
	}
}

func TestAuthHandler_UpdateProfile_422_EmptyName(t *testing.T) {
	h, authSvc, memberRepo, _ := setupAuthHandler()
	r := setupAuthRouter(h, authSvc, memberRepo)

	memberID := uuid.New()
	memberRepo.members[memberID] = &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true}

	cookie, _ := authSvc.CreateSessionCookie(memberID)

	body := `{"name":""}`
	req := httptest.NewRequest("PATCH", "/api/auth/profile", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}
