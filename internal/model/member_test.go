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

func TestMemberProfileFieldsSerialization(t *testing.T) {
	bio := "Full-stack developer passionate about Go"
	linkedin := "https://linkedin.com/in/testuser"
	instagram := "testuser"
	github := "testuser"
	now := time.Now().Truncate(time.Millisecond)

	m := Member{
		ID:              uuid.New(),
		Email:           "test@example.com",
		Name:            "Test User",
		IsActive:        true,
		Bio:             &bio,
		Skills:          []string{"go", "react", "postgresql"},
		LinkedinURL:     &linkedin,
		InstagramHandle: &instagram,
		GithubUsername:  &github,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal member: %v", err)
	}

	var decoded Member
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal member: %v", err)
	}

	if decoded.Bio == nil || *decoded.Bio != bio {
		t.Errorf("Bio = %v, want %q", decoded.Bio, bio)
	}
	if len(decoded.Skills) != 3 {
		t.Fatalf("Skills length = %d, want 3", len(decoded.Skills))
	}
	if decoded.Skills[0] != "go" || decoded.Skills[1] != "react" || decoded.Skills[2] != "postgresql" {
		t.Errorf("Skills = %v, want [go react postgresql]", decoded.Skills)
	}
	if decoded.LinkedinURL == nil || *decoded.LinkedinURL != linkedin {
		t.Errorf("LinkedinURL = %v, want %q", decoded.LinkedinURL, linkedin)
	}
	if decoded.InstagramHandle == nil || *decoded.InstagramHandle != instagram {
		t.Errorf("InstagramHandle = %v, want %q", decoded.InstagramHandle, instagram)
	}
	if decoded.GithubUsername == nil || *decoded.GithubUsername != github {
		t.Errorf("GithubUsername = %v, want %q", decoded.GithubUsername, github)
	}
}

func TestMemberProfileFieldsNullWhenUnset(t *testing.T) {
	m := Member{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		IsActive: true,
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if raw["bio"] != nil {
		t.Errorf("expected bio to be null, got %v", raw["bio"])
	}
	if raw["linkedin_url"] != nil {
		t.Errorf("expected linkedin_url to be null, got %v", raw["linkedin_url"])
	}
	if raw["instagram_handle"] != nil {
		t.Errorf("expected instagram_handle to be null, got %v", raw["instagram_handle"])
	}
	if raw["github_username"] != nil {
		t.Errorf("expected github_username to be null, got %v", raw["github_username"])
	}
	// Skills should be null when nil (not set), matching DB default behavior
	if raw["skills"] != nil {
		t.Errorf("expected skills to be null when nil, got %v", raw["skills"])
	}
}

func TestMemberEmptySkillsSerialization(t *testing.T) {
	m := Member{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		IsActive: true,
		Skills:   []string{},
	}

	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	skills, ok := raw["skills"].([]interface{})
	if !ok {
		t.Fatalf("expected skills to be an array, got %T", raw["skills"])
	}
	if len(skills) != 0 {
		t.Errorf("expected empty skills array, got %v", skills)
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
