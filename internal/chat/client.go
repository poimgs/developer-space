package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/service"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	maxMsgSize = 8192
	sendBufLen = 64
)

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	member  *model.Member
	send    chan []byte
	msgSvc  *service.MessageService
}

func NewClient(hub *Hub, conn *websocket.Conn, member *model.Member, msgSvc *service.MessageService) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		member: member,
		send:   make(chan []byte, sendBufLen),
		msgSvc: msgSvc,
	}
}

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	c.conn.SetReadLimit(maxMsgSize)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) != -1 {
				slog.Debug("ws connection closed", "member_id", c.member.ID, "status", websocket.CloseStatus(err))
			} else {
				slog.Debug("ws read error", "member_id", c.member.ID, "error", err)
			}
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError("Invalid message format")
			continue
		}

		switch msg.Type {
		case "send_message":
			c.handleSendMessage(ctx, msg.Payload)
		case "ping":
			c.sendJSON(WSMessage{Type: "pong", Payload: json.RawMessage(`{}`)})
		default:
			c.sendError("Unknown message type: " + msg.Type)
		}
	}
}

func (c *Client) WritePump(ctx context.Context) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Write(writeCtx, websocket.MessageText, message)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			pingCtx, cancel := context.WithTimeout(ctx, writeWait)
			err := c.conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) handleSendMessage(ctx context.Context, payload json.RawMessage) {
	var p SendMessagePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		c.sendError("Invalid send_message payload")
		return
	}

	channelID, err := uuid.Parse(p.ChannelID)
	if err != nil {
		c.sendError("Invalid channel_id")
		return
	}

	msg, err := c.msgSvc.Send(ctx, channelID, c.member.ID, p.Content)
	if err != nil {
		c.sendError("Failed to send message: " + err.Error())
		return
	}

	// Broadcast to all subscribers of this channel
	broadcast := WSMessage{Type: "message"}
	msgData, _ := json.Marshal(msg)
	broadcast.Payload = msgData
	c.hub.Broadcast(channelID, broadcast)
}

func (c *Client) sendError(msg string) {
	payload := ErrorPayload{Message: msg}
	c.sendJSON(WSMessage{Type: "error", Payload: mustMarshal(payload)})
}

func (c *Client) sendJSON(msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
		// Buffer full
	}
}

func mustMarshal(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

