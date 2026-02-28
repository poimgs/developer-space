package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

type mockCookieValidator struct {
	memberID uuid.UUID
	err      error
}

func (m *mockCookieValidator) ValidateSessionCookie(cookieValue string) (uuid.UUID, error) {
	if m.err != nil {
		return uuid.Nil, m.err
	}
	return m.memberID, nil
}

type mockMemberLookup struct {
	member *model.Member
	err    error
}

func (m *mockMemberLookup) GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.member, nil
}

func TestAuth_ValidCookie(t *testing.T) {
	memberID := uuid.New()
	member := &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: true, IsAdmin: false}

	cv := &mockCookieValidator{memberID: memberID}
	ml := &mockMemberLookup{member: member}

	var contextMember *model.Member
	handler := Auth(cv, ml)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextMember = MemberFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-cookie"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if contextMember == nil {
		t.Fatal("expected member in context")
	}
	if contextMember.ID != memberID {
		t.Errorf("expected member ID %s, got %s", memberID, contextMember.ID)
	}
}

func TestAuth_NoCookie(t *testing.T) {
	cv := &mockCookieValidator{}
	ml := &mockMemberLookup{}

	handler := Auth(cv, ml)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_InvalidCookie(t *testing.T) {
	cv := &mockCookieValidator{err: fmt.Errorf("decode error")}
	ml := &mockMemberLookup{}

	handler := Auth(cv, ml)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-value"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_InactiveMember(t *testing.T) {
	memberID := uuid.New()
	member := &model.Member{ID: memberID, Email: "jane@example.com", Name: "Jane", IsActive: false}

	cv := &mockCookieValidator{memberID: memberID}
	ml := &mockMemberLookup{member: member}

	handler := Auth(cv, ml)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for inactive member")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-cookie"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_MemberNotFound(t *testing.T) {
	cv := &mockCookieValidator{memberID: uuid.New()}
	ml := &mockMemberLookup{member: nil} // not found

	handler := Auth(cv, ml)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when member not found")
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-cookie"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMemberFromContext_NoMember(t *testing.T) {
	ctx := context.Background()
	member := MemberFromContext(ctx)
	if member != nil {
		t.Error("expected nil member from empty context")
	}
}
