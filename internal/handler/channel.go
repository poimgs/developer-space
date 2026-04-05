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

type ChannelHandler struct {
	svc *service.ChannelService
}

func NewChannelHandler(svc *service.ChannelService) *ChannelHandler {
	return &ChannelHandler{svc: svc}
}

func (h *ChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	channels, err := h.svc.List(r.Context())
	if err != nil {
		slog.Error("failed to list channels", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, channels)
}

func (h *ChannelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	channel, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			response.Error(w, http.StatusNotFound, "Channel not found")
			return
		}
		slog.Error("failed to get channel", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, channel)
}

func (h *ChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req model.CreateChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	channel, err := h.svc.Create(r.Context(), req, member.ID)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			response.ValidationError(w, ve.Details)
			return
		}
		slog.Error("failed to create channel", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusCreated, channel)
}

func (h *ChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid channel ID")
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrChannelNotFound) {
			response.Error(w, http.StatusNotFound, "Channel not found")
			return
		}
		if errors.Is(err, service.ErrCannotDeleteSession) {
			response.Error(w, http.StatusUnprocessableEntity, "Cannot delete a session channel")
			return
		}
		slog.Error("failed to delete channel", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
