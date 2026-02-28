package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTelegramService_Enabled(t *testing.T) {
	svc := NewTelegramService("bot123", "chat456")
	if !svc.Enabled() {
		t.Error("expected Enabled() = true when both token and chat ID set")
	}
}

func TestNewTelegramService_DisabledEmptyToken(t *testing.T) {
	svc := NewTelegramService("", "chat456")
	if svc.Enabled() {
		t.Error("expected Enabled() = false when token is empty")
	}
}

func TestNewTelegramService_DisabledEmptyChatID(t *testing.T) {
	svc := NewTelegramService("bot123", "")
	if svc.Enabled() {
		t.Error("expected Enabled() = false when chat ID is empty")
	}
}

func TestNewTelegramService_DisabledBothEmpty(t *testing.T) {
	svc := NewTelegramService("", "")
	if svc.Enabled() {
		t.Error("expected Enabled() = false when both are empty")
	}
}

func TestSendMessage_SkipsWhenDisabled(t *testing.T) {
	svc := NewTelegramService("", "")
	// Should not panic or error — just silently skip.
	svc.SendMessage("test message")
}

func TestSendMessage_SendsToAPI(t *testing.T) {
	var receivedBody sendMessageRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		// Verify the URL path contains the bot token
		expectedPath := "/bot123:ABC/sendMessage"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
		}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&receivedBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	svc := NewTelegramService("123:ABC", "chat789")
	svc.apiBase = server.URL

	svc.SendMessage("Hello World")

	if receivedBody.ChatID != "chat789" {
		t.Errorf("expected chat_id=chat789, got %s", receivedBody.ChatID)
	}
	if receivedBody.Text != "Hello World" {
		t.Errorf("expected text=Hello World, got %s", receivedBody.Text)
	}
	if receivedBody.ParseMode != "MarkdownV2" {
		t.Errorf("expected parse_mode=MarkdownV2, got %s", receivedBody.ParseMode)
	}
}

func TestSendMessage_HandlesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false,"description":"Bad Request"}`))
	}))
	defer server.Close()

	svc := NewTelegramService("bot123", "chat456")
	svc.apiBase = server.URL

	// Should not panic — just logs a warning.
	svc.SendMessage("test")
}

func TestSendMessage_HandlesNetworkError(t *testing.T) {
	svc := NewTelegramService("bot123", "chat456")
	svc.apiBase = "http://localhost:1" // unreachable

	// Should not panic — just logs a warning.
	svc.SendMessage("test")
}
