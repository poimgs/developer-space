package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSpaceSessionJSONSerialization(t *testing.T) {
	desc := "A collaborative session"
	now := time.Now().Truncate(time.Millisecond)
	s := SpaceSession{
		ID:          uuid.New(),
		Title:       "Morning Session",
		Description: &desc,
		Date:        "2026-03-15",
		StartTime:   "09:00",
		EndTime:     "12:00",
		Capacity:    8,
		Status:      "scheduled",
		CreatedBy:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		RSVPCount:   3,
		UserRSVPed:  true,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal session: %v", err)
	}

	var decoded SpaceSession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal session: %v", err)
	}

	if decoded.Title != s.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, s.Title)
	}
	if decoded.Description == nil || *decoded.Description != desc {
		t.Errorf("Description = %v, want %q", decoded.Description, desc)
	}
	if decoded.Date != s.Date {
		t.Errorf("Date = %q, want %q", decoded.Date, s.Date)
	}
	if decoded.Capacity != 8 {
		t.Errorf("Capacity = %d, want 8", decoded.Capacity)
	}
	if decoded.Status != "scheduled" {
		t.Errorf("Status = %q, want 'scheduled'", decoded.Status)
	}
	if decoded.RSVPCount != 3 {
		t.Errorf("RSVPCount = %d, want 3", decoded.RSVPCount)
	}
	if !decoded.UserRSVPed {
		t.Error("UserRSVPed should be true")
	}
}

func TestSpaceSessionOmitsZeroRSVPCount(t *testing.T) {
	s := SpaceSession{
		ID:       uuid.New(),
		Title:    "Test",
		Date:     "2026-03-15",
		Capacity: 5,
		Status:   "scheduled",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := raw["rsvp_count"]; ok {
		t.Error("rsvp_count should be omitted when zero")
	}
}

func TestCreateSessionRequestJSON(t *testing.T) {
	body := `{"title":"Afternoon Session","date":"2026-03-20","start_time":"14:00","end_time":"17:00","capacity":6,"repeat_weekly":3}`

	var req CreateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title != "Afternoon Session" {
		t.Errorf("Title = %q, want 'Afternoon Session'", req.Title)
	}
	if req.Capacity != 6 {
		t.Errorf("Capacity = %d, want 6", req.Capacity)
	}
	if req.RepeatWeekly != 3 {
		t.Errorf("RepeatWeekly = %d, want 3", req.RepeatWeekly)
	}
}

func TestUpdateSessionRequestPartialJSON(t *testing.T) {
	body := `{"title":"Updated Title","capacity":10}`

	var req UpdateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title == nil || *req.Title != "Updated Title" {
		t.Errorf("Title = %v, want 'Updated Title'", req.Title)
	}
	if req.Capacity == nil || *req.Capacity != 10 {
		t.Errorf("Capacity = %v, want 10", req.Capacity)
	}
	if req.Date != nil {
		t.Error("Date should be nil for partial update")
	}
	if req.StartTime != nil {
		t.Error("StartTime should be nil for partial update")
	}
}
