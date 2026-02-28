package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/response"
)

const MemberKey contextKey = "member"

// CookieValidator decodes session cookies and returns member IDs.
type CookieValidator interface {
	ValidateSessionCookie(cookieValue string) (uuid.UUID, error)
}

// MemberLookup fetches a member by ID.
type MemberLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*model.Member, error)
}

// Auth middleware validates the session cookie, loads the member, and stores it in context.
func Auth(cv CookieValidator, ml MemberLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Not authenticated")
				return
			}

			memberID, err := cv.ValidateSessionCookie(cookie.Value)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "Not authenticated")
				return
			}

			member, err := ml.GetByID(r.Context(), memberID)
			if err != nil || member == nil || !member.IsActive {
				response.Error(w, http.StatusUnauthorized, "Not authenticated")
				return
			}

			ctx := context.WithValue(r.Context(), MemberKey, member)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MemberFromContext extracts the authenticated member from the request context.
func MemberFromContext(ctx context.Context) *model.Member {
	if m, ok := ctx.Value(MemberKey).(*model.Member); ok {
		return m
	}
	return nil
}
