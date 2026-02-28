package middleware

import (
	"net/http"

	"github.com/developer-space/api/internal/response"
)

// Admin middleware checks that the authenticated member is an admin.
// Must be applied after Auth middleware.
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		member := MemberFromContext(r.Context())
		if member == nil || !member.IsAdmin {
			response.Error(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
