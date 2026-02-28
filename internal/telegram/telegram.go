package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// TelegramService sends messages to a Telegram group chat via the Bot API.
// If bot token or chat ID are empty, all operations are silently skipped.
type TelegramService struct {
	botToken string
	chatID   string
	client   *http.Client
	apiBase  string // overridable for testing; defaults to Telegram API
}

// NewTelegramService creates a new TelegramService.
// If botToken or chatID are empty, the service operates as a no-op.
func NewTelegramService(botToken, chatID string) *TelegramService {
	return &TelegramService{
		botToken: botToken,
		chatID:   chatID,
		client:   &http.Client{},
		apiBase:  "https://api.telegram.org",
	}
}

// Enabled returns true if Telegram credentials are configured.
func (s *TelegramService) Enabled() bool {
	return s.botToken != "" && s.chatID != ""
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// SendMessage sends a MarkdownV2-formatted message to the configured chat.
// Returns nil immediately if credentials are not configured.
// Failures are logged at warn level but never returned to the caller.
func (s *TelegramService) SendMessage(text string) {
	if !s.Enabled() {
		slog.Debug("telegram notification skipped: credentials not configured")
		return
	}

	payload := sendMessageRequest{
		ChatID:    s.chatID,
		Text:      text,
		ParseMode: "MarkdownV2",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Warn("telegram: failed to marshal message", "error", err)
		return
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", s.apiBase, s.botToken)
	resp, err := s.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		slog.Warn("telegram: failed to send message", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Warn("telegram: API returned error", "status", resp.StatusCode)
		return
	}

	slog.Debug("telegram: message sent successfully")
}
