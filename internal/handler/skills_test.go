package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/service"
)

func setupSkillsHandler() (*SkillsHandler, *mockMemberRepo) {
	repo := newMockRepo()
	svc := service.NewMemberService(repo, &noopEmailSender{}, "http://localhost:5173")
	h := NewSkillsHandler(svc)
	return h, repo
}

func setupSkillsRouter(h *SkillsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/skills", h.List)
	return r
}

func TestSkillsHandler_List_200_Empty(t *testing.T) {
	h, _ := setupSkillsHandler()
	r := setupSkillsRouter(h)

	req := httptest.NewRequest("GET", "/api/skills", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data []string `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data == nil {
		t.Error("expected non-nil data array")
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected empty skills, got %d", len(resp.Data))
	}
}

func TestSkillsHandler_List_200_WithSkills(t *testing.T) {
	h, repo := setupSkillsHandler()
	r := setupSkillsRouter(h)

	id := uuid.New()
	repo.members[id] = &model.Member{
		ID:       id,
		Email:    "alice@example.com",
		Name:     "Alice",
		IsActive: true,
		Skills:   []string{"go", "react"},
	}

	req := httptest.NewRequest("GET", "/api/skills", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Data []string `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 skills, got %d", len(resp.Data))
	}
}
