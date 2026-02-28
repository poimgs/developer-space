package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type MemberHandler struct {
	svc *service.MemberService
}

func NewMemberHandler(svc *service.MemberService) *MemberHandler {
	return &MemberHandler{svc: svc}
}

func (h *MemberHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	member, err := h.svc.Create(r.Context(), req)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			response.ValidationError(w, ve.Details)
			return
		}
		if errors.Is(err, service.ErrDuplicateEmail) {
			response.Error(w, http.StatusConflict, "A member with this email already exists")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, member)
}

func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	activeFilter := r.URL.Query().Get("active")
	if activeFilter == "" {
		activeFilter = "true"
	}

	members, err := h.svc.List(r.Context(), activeFilter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, members)
}

func (h *MemberHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	member, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Member not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, member)
}

func (h *MemberHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	var req model.UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	member, err := h.svc.Update(r.Context(), id, req)
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

	response.JSON(w, http.StatusOK, member)
}

func (h *MemberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Error(w, http.StatusNotFound, "Member not found")
			return
		}
		if errors.Is(err, service.ErrHasRSVPs) {
			response.Error(w, http.StatusConflict, "Cannot delete member with existing RSVPs. Deactivate instead.")
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
