package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type ProfileHandler struct {
	svc *service.AuthService
}

func NewProfileHandler(svc *service.AuthService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	profile, err := h.svc.GetPublicProfile(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Member not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, profile)
}
