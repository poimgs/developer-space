package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/developer-space/api/internal/model"
	"github.com/google/uuid"
)

// captureSvc creates a TelegramService that captures the sent message text.
func captureSvc(t *testing.T) (*TelegramService, *string) {
	t.Helper()
	var captured string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		captured = body.Text
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)

	svc := NewTelegramService("testtoken", "testchat")
	svc.apiBase = server.URL
	return svc, &captured
}

func testSession() *model.SpaceSession {
	desc := "Open co-working, bring your laptop"
	return &model.SpaceSession{
		ID:          uuid.New(),
		Title:       "Friday Afternoon Session",
		Description: &desc,
		Date:        "2025-02-14",
		StartTime:   "14:00",
		EndTime:     "18:00",
		Capacity:    8,
		Status:      "scheduled",
		RSVPCount:   4,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func testMember() *model.Member {
	return &model.Member{
		ID:    uuid.New(),
		Email: "jane@example.com",
		Name:  "Jane Doe",
	}
}

func TestNotifier_SessionCreated(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	notifier.SessionCreated(session)

	assertContains(t, *captured, "New Session")
	assertContains(t, *captured, "Friday Afternoon Session")
	assertContains(t, *captured, "2025\\-02\\-14")
	assertContains(t, *captured, "14:00")
	assertContains(t, *captured, "18:00")
	assertContains(t, *captured, "8 spots available")
	assertContains(t, *captured, "Open co\\-working")
}

func TestNotifier_SessionCreatedNoDescription(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	session.Description = nil
	notifier.SessionCreated(session)

	assertContains(t, *captured, "New Session")
	assertNotContains(t, *captured, "laptop")
}

func TestNotifier_SessionsCreatedRecurring(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	sessions := []model.SpaceSession{
		{Title: "Friday Session", Date: "2025-02-14", StartTime: "14:00", EndTime: "18:00", Capacity: 8},
		{Title: "Friday Session", Date: "2025-02-21", StartTime: "14:00", EndTime: "18:00", Capacity: 8},
		{Title: "Friday Session", Date: "2025-02-28", StartTime: "14:00", EndTime: "18:00", Capacity: 8},
	}
	notifier.SessionsCreatedRecurring(sessions)

	assertContains(t, *captured, "Recurring Sessions Created")
	assertContains(t, *captured, "Friday Session")
	assertContains(t, *captured, "8 spots each")
	assertContains(t, *captured, "2025\\-02\\-14")
	assertContains(t, *captured, "2025\\-02\\-21")
	assertContains(t, *captured, "2025\\-02\\-28")
}

func TestNotifier_SessionsCreatedRecurringEmpty(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	notifier.SessionsCreatedRecurring([]model.SpaceSession{})

	if *captured != "" {
		t.Errorf("expected no message for empty sessions, got %q", *captured)
	}
}

func TestNotifier_SessionShifted(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	session.Date = "2025-02-15"
	session.StartTime = "15:00"
	session.EndTime = "19:00"
	session.Status = "shifted"
	notifier.SessionShifted(session)

	assertContains(t, *captured, "Session Rescheduled")
	assertContains(t, *captured, "Friday Afternoon Session")
	assertContains(t, *captured, "2025\\-02\\-15")
	assertContains(t, *captured, "15:00")
	assertContains(t, *captured, "19:00")
}

func TestNotifier_SessionCanceled(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	notifier.SessionCanceled(session)

	assertContains(t, *captured, "Session Canceled")
	assertContains(t, *captured, "Friday Afternoon Session")
	assertContains(t, *captured, "2025\\-02\\-14")
	assertContains(t, *captured, "14:00")
	assertContains(t, *captured, "18:00")
}

func TestNotifier_MemberRSVPed(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	session.RSVPCount = 4
	member := testMember()
	notifier.MemberRSVPed(session, member)

	assertContains(t, *captured, "Jane Doe RSVPed")
	assertContains(t, *captured, "Friday Afternoon Session")
	assertContains(t, *captured, "4 / 8 spots taken")
}

func TestNotifier_MemberCanceledRSVP(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	session.RSVPCount = 3
	member := testMember()
	notifier.MemberCanceledRSVP(session, member)

	assertContains(t, *captured, "Jane Doe canceled RSVP")
	assertContains(t, *captured, "Friday Afternoon Session")
	assertContains(t, *captured, "3 / 8 spots taken")
}

func TestNotifier_EscapesSpecialChars(t *testing.T) {
	svc, captured := captureSvc(t)
	notifier := NewTelegramNotifier(svc)

	session := testSession()
	session.Title = "Test_Session*Special.Chars!"
	notifier.SessionCanceled(session)

	assertContains(t, *captured, `Test\_Session\*Special\.Chars\!`)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected message to contain %q, got:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected message NOT to contain %q, got:\n%s", substr, s)
	}
}
