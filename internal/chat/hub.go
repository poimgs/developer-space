package chat

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

type subscription struct {
	client *Client
	done   bool // true = unregister
}

type broadcastMsg struct {
	channelID uuid.UUID
	data      []byte
	sender    *Client // nil means send to everyone including sender
}

type Hub struct {
	mu         sync.RWMutex
	clients    map[*Client]struct{}
	rooms      map[uuid.UUID]map[*Client]struct{}
	register   chan *subscription
	unregister chan *subscription
	broadcast  chan *broadcastMsg
	done       chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		rooms:      make(map[uuid.UUID]map[*Client]struct{}),
		register:   make(chan *subscription, 64),
		unregister: make(chan *subscription, 64),
		broadcast:  make(chan *broadcastMsg, 256),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.register:
			h.mu.Lock()
			h.clients[sub.client] = struct{}{}
			h.mu.Unlock()
			slog.Debug("ws client registered", "member_id", sub.client.member.ID)

		case sub := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[sub.client]; ok {
				delete(h.clients, sub.client)
				// Remove from all rooms
				for chID, room := range h.rooms {
					delete(room, sub.client)
					if len(room) == 0 {
						delete(h.rooms, chID)
					}
				}
				close(sub.client.send)
			}
			h.mu.Unlock()
			slog.Debug("ws client unregistered", "member_id", sub.client.member.ID)

		case msg := <-h.broadcast:
			h.mu.RLock()
			if room, ok := h.rooms[msg.channelID]; ok {
				for client := range room {
					select {
					case client.send <- msg.data:
					default:
						// Client send buffer full, disconnect
						go h.Unregister(client)
					}
				}
			}
			h.mu.RUnlock()

		case <-h.done:
			return
		}
	}
}

func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) Register(c *Client) {
	h.register <- &subscription{client: c}
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- &subscription{client: c, done: true}
}

func (h *Hub) Subscribe(c *Client, channelID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[channelID] == nil {
		h.rooms[channelID] = make(map[*Client]struct{})
	}
	h.rooms[channelID][c] = struct{}{}
}

func (h *Hub) SubscribeAll(c *Client, channelIDs []uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, chID := range channelIDs {
		if h.rooms[chID] == nil {
			h.rooms[chID] = make(map[*Client]struct{})
		}
		h.rooms[chID][c] = struct{}{}
	}
}

func (h *Hub) Broadcast(channelID uuid.UUID, msg any) {
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal broadcast message", "error", err)
		return
	}
	h.broadcast <- &broadcastMsg{
		channelID: channelID,
		data:      data,
	}
}
