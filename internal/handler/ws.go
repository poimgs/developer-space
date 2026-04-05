package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/chat"
	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

type WSHandler struct {
	hub     *chat.Hub
	chSvc   *service.ChannelService
	msgSvc  *service.MessageService
	origins []string
}

func NewWSHandler(hub *chat.Hub, chSvc *service.ChannelService, msgSvc *service.MessageService, frontendURL string) *WSHandler {
	return &WSHandler{
		hub:     hub,
		chSvc:   chSvc,
		msgSvc:  msgSvc,
		origins: []string{frontendURL},
	}
}

func (h *WSHandler) Connect(w http.ResponseWriter, r *http.Request) {
	member := middleware.MemberFromContext(r.Context())
	if member == nil {
		response.Error(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	// Clear the write deadline so the HTTP server's WriteTimeout doesn't kill WS
	rc := http.NewResponseController(w)
	if err := rc.SetWriteDeadline(time.Time{}); err != nil {
		slog.Warn("failed to clear write deadline", "error", err)
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: h.origins,
	})
	if err != nil {
		slog.Error("failed to accept websocket", "error", err)
		return
	}

	// Get all channels so client can subscribe to all
	channels, err := h.chSvc.List(r.Context())
	if err != nil {
		slog.Error("failed to list channels for ws subscription", "error", err)
		conn.Close(websocket.StatusInternalError, "Failed to load channels")
		return
	}

	client := chat.NewClient(h.hub, conn, member, h.msgSvc)
	h.hub.Register(client)

	// Subscribe to all channels
	var channelIDs []uuid.UUID
	for _, ch := range channels {
		channelIDs = append(channelIDs, ch.ID)
	}
	h.hub.SubscribeAll(client, channelIDs)

	ctx := r.Context()
	go client.WritePump(ctx)
	go client.ReadPump(ctx)
}
