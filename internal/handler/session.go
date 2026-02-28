package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type SessionHandler struct {
	svc *service.SessionService
}

func NewSessionHandler(svc *service.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req model.CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.svc.Create(r.Context(), req, member.ID)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			response.ValidationError(w, ve.Details)
			return
		}
		slog.Error("failed to create session", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, result)
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	status := r.URL.Query().Get("status")

	var memberID *uuid.UUID
	if member := middleware.MemberFromContext(r.Context()); member != nil {
		memberID = &member.ID
	}

	sessions, err := h.svc.List(r.Context(), from, to, status, memberID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, sessions)
}

func (h *SessionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var memberID *uuid.UUID
	if member := middleware.MemberFromContext(r.Context()); member != nil {
		memberID = &member.ID
	}

	session, err := h.svc.GetByID(r.Context(), id, memberID)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	var req model.UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	session, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			response.ValidationError(w, ve.Details)
			return
		}
		if errors.Is(err, service.ErrSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, service.ErrSessionCanceled) {
			response.Error(w, http.StatusUnprocessableEntity, "Cannot edit a canceled session")
			return
		}
		if errors.Is(err, service.ErrCapacityBelowRSVPs) {
			response.Error(w, http.StatusConflict, "Cannot reduce capacity below current RSVP count")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, session)
}

func (h *SessionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	session, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, service.ErrAlreadyCanceled) {
			response.Error(w, http.StatusUnprocessableEntity, "Session is already canceled")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, session)
}
