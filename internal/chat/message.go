package chat

import "encoding/json"

// WSMessage is the envelope for all WebSocket communication.
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// SendMessagePayload is sent by clients to post a message.
type SendMessagePayload struct {
	ChannelID string `json:"channel_id"`
	Content   string `json:"content"`
}

// ErrorPayload is sent by the server on errors.
type ErrorPayload struct {
	Message string `json:"message"`
}
