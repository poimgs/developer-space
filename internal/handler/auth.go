package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type magicLinkRequest struct {
	Email string `json:"email"`
}

type messageResponse struct {
	Message string `json:"message"`
}

func (h *AuthHandler) RequestMagicLink(w http.ResponseWriter, r *http.Request) {
	var req magicLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.svc.RequestMagicLink(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, service.ErrRateLimited) {
			response.Error(w, http.StatusTooManyRequests, "Too many requests. Try again later.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, messageResponse{
		Message: "If this email is registered, a login link has been sent.",
	})
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		response.Error(w, http.StatusBadRequest, "Token is required")
		return
	}

	member, err := h.svc.VerifyToken(r.Context(), token)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			response.Error(w, http.StatusUnauthorized, "Invalid or expired login link.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	cookie, err := h.svc.CreateSessionCookie(member.ID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	response.JSON(w, http.StatusOK, member)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.svc.ClearSessionCookie())
	response.JSON(w, http.StatusOK, messageResponse{Message: "Logged out"})
}

type updateProfileRequest struct {
	Name           *string `json:"name"`
	TelegramHandle *string `json:"telegram_handle"`
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updated, err := h.svc.UpdateProfile(r.Context(), member.ID, req.Name, req.TelegramHandle)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			response.ValidationError(w, ve.Details)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Member not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, updated)
}
