package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMemberJSONSerialization(t *testing.T) {
	handle := "@testuser"
	now := time.Now().Truncate(time.Millisecond)
	m := Member{
		ID:             uuid.New(),
		Email:          "test@example.com",
		Name:           "Test User",
		TelegramHandle: &handle,
		IsAdmin:        true,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal member: %v", err)
	}

	var decoded Member
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal member: %v", err)
	}

	if decoded.ID != m.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, m.ID)
	}
	if decoded.Email != m.Email {
		t.Errorf("Email = %q, want %q", decoded.Email, m.Email)
	}
	if decoded.Name != m.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, m.Name)
	}
	if decoded.TelegramHandle == nil || *decoded.TelegramHandle != handle {
		t.Errorf("TelegramHandle = %v, want %q", decoded.TelegramHandle, handle)
	}
	if decoded.IsAdmin != m.IsAdmin {
		t.Errorf("IsAdmin = %v, want %v", decoded.IsAdmin, m.IsAdmin)
	}
	if decoded.IsActive != m.IsActive {
		t.Errorf("IsActive = %v, want %v", decoded.IsActive, m.IsActive)
	}
}

func TestMemberNilTelegramHandle(t *testing.T) {
	m := Member{
		ID:    uuid.New(),
		Email: "test@example.com",
		Name:  "Test User",
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if raw["telegram_handle"] != nil {
		t.Errorf("expected telegram_handle to be null, got %v", raw["telegram_handle"])
	}
}

func TestCreateMemberRequestJSON(t *testing.T) {
	body := `{"email":"alice@example.com","name":"Alice","telegram_handle":"@alice","is_admin":false,"send_invite":true}`

	var req CreateMemberRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", req.Email, "alice@example.com")
	}
	if req.Name != "Alice" {
		t.Errorf("Name = %q, want %q", req.Name, "Alice")
	}
	if req.TelegramHandle == nil || *req.TelegramHandle != "@alice" {
		t.Errorf("TelegramHandle = %v, want @alice", req.TelegramHandle)
	}
	if req.IsAdmin {
		t.Error("IsAdmin should be false")
	}
	if !req.SendInvite {
		t.Error("SendInvite should be true")
	}
}

func TestUpdateMemberRequestPartialJSON(t *testing.T) {
	body := `{"name":"New Name"}`

	var req UpdateMemberRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Name == nil || *req.Name != "New Name" {
		t.Errorf("Name = %v, want 'New Name'", req.Name)
	}
	if req.TelegramHandle != nil {
		t.Error("TelegramHandle should be nil for partial update")
	}
	if req.IsAdmin != nil {
		t.Error("IsAdmin should be nil for partial update")
	}
	if req.IsActive != nil {
		t.Error("IsActive should be nil for partial update")
	}
}
