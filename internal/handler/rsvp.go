package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type RSVPHandler struct {
	svc *service.RSVPService
}

func NewRSVPHandler(svc *service.RSVPService) *RSVPHandler {
	return &RSVPHandler{svc: svc}
}

func (h *RSVPHandler) RSVP(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rsvp, err := h.svc.Create(r.Context(), sessionID, member.ID)
	if err != nil {
		if errors.Is(err, service.ErrRSVPSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, service.ErrRSVPSessionCanceled) {
			response.Error(w, http.StatusUnprocessableEntity, "Cannot RSVP to a canceled session")
			return
		}
		if errors.Is(err, service.ErrRSVPSessionPast) {
			response.Error(w, http.StatusUnprocessableEntity, "Cannot RSVP to a past session")
			return
		}
		if errors.Is(err, service.ErrRSVPDuplicate) {
			response.Error(w, http.StatusConflict, "You have already RSVPed to this session")
			return
		}
		if errors.Is(err, service.ErrRSVPSessionFull) {
			response.Error(w, http.StatusConflict, "This session is full")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, rsvp)
}

func (h *RSVPHandler) CancelRSVP(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	err = h.svc.Cancel(r.Context(), sessionID, member.ID)
	if err != nil {
		if errors.Is(err, service.ErrRSVPSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		if errors.Is(err, service.ErrRSVPSessionPast) {
			response.Error(w, http.StatusUnprocessableEntity, "Cannot cancel RSVP for a past session")
			return
		}
		if errors.Is(err, service.ErrRSVPNotFound) {
			response.Error(w, http.StatusNotFound, "You have not RSVPed to this session")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RSVPHandler) ListRSVPs(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	rsvps, err := h.svc.ListBySession(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, service.ErrRSVPSessionNotFound) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, rsvps)
}
