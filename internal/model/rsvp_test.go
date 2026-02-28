package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRSVPJSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	r := RSVP{
		ID:        uuid.New(),
		SessionID: uuid.New(),
		MemberID:  uuid.New(),
		CreatedAt: now,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal RSVP: %v", err)
	}

	var decoded RSVP
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal RSVP: %v", err)
	}

	if decoded.ID != r.ID {
		t.Errorf("ID = %v, want %v", decoded.ID, r.ID)
	}
	if decoded.SessionID != r.SessionID {
		t.Errorf("SessionID = %v, want %v", decoded.SessionID, r.SessionID)
	}
	if decoded.MemberID != r.MemberID {
		t.Errorf("MemberID = %v, want %v", decoded.MemberID, r.MemberID)
	}
}

func TestRSVPWithMemberJSONSerialization(t *testing.T) {
	handle := "@alice"
	now := time.Now().Truncate(time.Millisecond)
	r := RSVPWithMember{
		ID:        uuid.New(),
		SessionID: uuid.New(),
		Member: RSVPMember{
			ID:             uuid.New(),
			Name:           "Alice",
			TelegramHandle: &handle,
		},
		CreatedAt: now,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded RSVPWithMember
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Member.Name != "Alice" {
		t.Errorf("Member.Name = %q, want 'Alice'", decoded.Member.Name)
	}
	if decoded.Member.TelegramHandle == nil || *decoded.Member.TelegramHandle != "@alice" {
		t.Errorf("Member.TelegramHandle = %v, want '@alice'", decoded.Member.TelegramHandle)
	}
}
