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
		Status:      "scheduled",
		CreatedBy:   uuid.New(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Capacity:    20,
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
		ID:     uuid.New(),
		Title:  "Test",
		Date:   "2026-03-15",
		Status: "scheduled",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := raw["rsvp_count"]; !ok {
		t.Error("rsvp_count should be present even when zero")
	}
	if raw["rsvp_count"].(float64) != 0 {
		t.Error("rsvp_count should be 0")
	}
}

func TestSpaceSessionImageURLAndLocation(t *testing.T) {
	imageURL := "/uploads/sessions/abc-123.jpg"
	location := "WeWork Coworking Space, Floor 3"
	now := time.Now().Truncate(time.Millisecond)

	s := SpaceSession{
		ID:        uuid.New(),
		Title:     "Session with Image",
		Date:      "2026-03-15",
		StartTime: "09:00",
		EndTime:   "12:00",
		Status:    "scheduled",
		ImageURL:  &imageURL,
		Location:  &location,
		CreatedBy: uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal session: %v", err)
	}

	var decoded SpaceSession
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal session: %v", err)
	}

	if decoded.ImageURL == nil || *decoded.ImageURL != imageURL {
		t.Errorf("ImageURL = %v, want %q", decoded.ImageURL, imageURL)
	}
	if decoded.Location == nil || *decoded.Location != location {
		t.Errorf("Location = %v, want %q", decoded.Location, location)
	}
}

func TestSpaceSessionImageURLAndLocationNullWhenUnset(t *testing.T) {
	s := SpaceSession{
		ID:     uuid.New(),
		Title:  "Session without Image",
		Date:   "2026-03-15",
		Status: "scheduled",
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if raw["image_url"] != nil {
		t.Errorf("expected image_url to be null, got %v", raw["image_url"])
	}
	if raw["location"] != nil {
		t.Errorf("expected location to be null, got %v", raw["location"])
	}
}

func TestCreateSessionRequestWithLocation(t *testing.T) {
	body := `{"title":"Afternoon Session","date":"2026-03-20","start_time":"14:00","end_time":"17:00","location":"Downtown Office"}`

	var req CreateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title != "Afternoon Session" {
		t.Errorf("Title = %q, want 'Afternoon Session'", req.Title)
	}
	if req.Location == nil || *req.Location != "Downtown Office" {
		t.Errorf("Location = %v, want 'Downtown Office'", req.Location)
	}
}

func TestCreateSessionRequestLocationOptional(t *testing.T) {
	body := `{"title":"Afternoon Session","date":"2026-03-20","start_time":"14:00","end_time":"17:00"}`

	var req CreateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Location != nil {
		t.Errorf("Location should be nil when not provided, got %v", req.Location)
	}
}

func TestUpdateSessionRequestWithLocation(t *testing.T) {
	body := `{"location":"New Location"}`

	var req UpdateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Location == nil || *req.Location != "New Location" {
		t.Errorf("Location = %v, want 'New Location'", req.Location)
	}
	if req.Title != nil {
		t.Error("Title should be nil for partial update")
	}
}

func TestCreateSessionRequestJSON(t *testing.T) {
	body := `{"title":"Afternoon Session","date":"2026-03-20","start_time":"14:00","end_time":"17:00","repeat_weekly":3}`

	var req CreateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title != "Afternoon Session" {
		t.Errorf("Title = %q, want 'Afternoon Session'", req.Title)
	}
	if req.RepeatWeekly != 3 {
		t.Errorf("RepeatWeekly = %d, want 3", req.RepeatWeekly)
	}
}

func TestUpdateSessionRequestPartialJSON(t *testing.T) {
	body := `{"title":"Updated Title"}`

	var req UpdateSessionRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Title == nil || *req.Title != "Updated Title" {
		t.Errorf("Title = %v, want 'Updated Title'", req.Title)
	}
	if req.Date != nil {
		t.Error("Date should be nil for partial update")
	}
	if req.StartTime != nil {
		t.Error("StartTime should be nil for partial update")
	}
}
