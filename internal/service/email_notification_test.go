package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

// --- Mock implementations ---

type sentEmail struct {
	To      string
	Subject string
	Body    string
}

type mockNotificationEmailSender struct {
	mu    sync.Mutex
	sent  []sentEmail
}

func (m *mockNotificationEmailSender) SendNotificationEmail(ctx context.Context, to, subject, body string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, sentEmail{To: to, Subject: subject, Body: body})
}

func (m *mockNotificationEmailSender) getSent() []sentEmail {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]sentEmail, len(m.sent))
	copy(result, m.sent)
	return result
}

type mockRSVPMemberLister struct {
	recipients map[uuid.UUID][]model.RSVPRecipient
	err        error
}

func (m *mockRSVPMemberLister) ListEmailsBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPRecipient, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.recipients[sessionID], nil
}

// --- Email Template Tests ---

func TestFormatDateHuman(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-02-14", "February 14, 2025"},
		{"2025-12-25", "December 25, 2025"},
		{"2025-01-01", "January 1, 2025"},
		{"bad-date", "bad-date"},
	}

	for _, tt := range tests {
		got := formatDateHuman(tt.input)
		if got != tt.expected {
			t.Errorf("formatDateHuman(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestCancelEmailSubject(t *testing.T) {
	subject := cancelEmailSubject("Friday Afternoon Session")
	expected := "Session Canceled: Friday Afternoon Session"
	if subject != expected {
		t.Errorf("got %q, want %q", subject, expected)
	}
}

func TestCancelEmailBody(t *testing.T) {
	session := &model.SpaceSession{
		Title:     "Friday Afternoon Session",
		Date:      "2025-02-14",
		StartTime: "14:00",
		EndTime:   "18:00",
	}

	body := cancelEmailBody("Jane", session)

	expectedParts := []string{
		"Hi Jane,",
		`"Friday Afternoon Session"`,
		"February 14, 2025",
		"14:00 – 18:00",
		"has been canceled",
		"— Developer Space",
	}

	for _, part := range expectedParts {
		if !containsStr(body, part) {
			t.Errorf("cancel email body missing %q\ngot: %s", part, body)
		}
	}
}

func TestRescheduleEmailSubject(t *testing.T) {
	subject := rescheduleEmailSubject("Friday Afternoon Session")
	expected := "Session Rescheduled: Friday Afternoon Session"
	if subject != expected {
		t.Errorf("got %q, want %q", subject, expected)
	}
}

func TestRescheduleEmailBody(t *testing.T) {
	oldSession := &model.SpaceSession{
		Title:     "Friday Afternoon Session",
		Date:      "2025-02-14",
		StartTime: "14:00",
		EndTime:   "18:00",
	}
	newSession := &model.SpaceSession{
		Title:     "Friday Afternoon Session",
		Date:      "2025-02-15",
		StartTime: "15:00",
		EndTime:   "19:00",
	}

	body := rescheduleEmailBody("Jane", oldSession, newSession)

	expectedParts := []string{
		"Hi Jane,",
		`"Friday Afternoon Session"`,
		"Previously: February 14, 2025, 14:00 – 18:00",
		"Now:        February 15, 2025, 15:00 – 19:00",
		"RSVP is still active",
		"— Developer Space",
	}

	for _, part := range expectedParts {
		if !containsStr(body, part) {
			t.Errorf("reschedule email body missing %q\ngot: %s", part, body)
		}
	}
}

// --- sendCancelEmails Tests ---

func TestSendCancelEmails_SendsToAllRecipients(t *testing.T) {
	repo := newMockSessionRepo()
	emailSender := &mockNotificationEmailSender{}
	sessionID := uuid.New()
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			sessionID: {
				{Name: "Alice", Email: "alice@example.com"},
				{Name: "Bob", Email: "bob@example.com"},
			},
		},
	}

	svc := NewSessionService(repo, &mockNotifier{})
	svc.SetEmailNotifier(emailSender, rsvpLister)

	session := &model.SpaceSession{
		ID:        sessionID,
		Title:     "Test Session",
		Date:      "2025-03-01",
		StartTime: "10:00",
		EndTime:   "14:00",
	}

	svc.sendCancelEmails(session)

	sent := emailSender.getSent()
	if len(sent) != 2 {
		t.Fatalf("expected 2 emails sent, got %d", len(sent))
	}

	if sent[0].To != "alice@example.com" {
		t.Errorf("expected first email to alice, got %q", sent[0].To)
	}
	if sent[1].To != "bob@example.com" {
		t.Errorf("expected second email to bob, got %q", sent[1].To)
	}

	// Verify subject
	expectedSubject := "Session Canceled: Test Session"
	if sent[0].Subject != expectedSubject {
		t.Errorf("got subject %q, want %q", sent[0].Subject, expectedSubject)
	}

	// Verify body contains member name
	if !containsStr(sent[0].Body, "Hi Alice,") {
		t.Error("first email should address Alice")
	}
	if !containsStr(sent[1].Body, "Hi Bob,") {
		t.Error("second email should address Bob")
	}
}

func TestSendCancelEmails_NoRecipients(t *testing.T) {
	repo := newMockSessionRepo()
	emailSender := &mockNotificationEmailSender{}
	sessionID := uuid.New()
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			sessionID: {},
		},
	}

	svc := NewSessionService(repo, &mockNotifier{})
	svc.SetEmailNotifier(emailSender, rsvpLister)

	session := &model.SpaceSession{
		ID:        sessionID,
		Title:     "Empty Session",
		Date:      "2025-03-01",
		StartTime: "10:00",
		EndTime:   "14:00",
	}

	svc.sendCancelEmails(session)

	sent := emailSender.getSent()
	if len(sent) != 0 {
		t.Errorf("expected 0 emails sent, got %d", len(sent))
	}
}

func TestSendCancelEmails_NilDependencies(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})
	// No SetEmailNotifier called — emailSender and rsvpLister are nil

	session := &model.SpaceSession{
		ID:    uuid.New(),
		Title: "Test",
	}

	// Should not panic
	svc.sendCancelEmails(session)
}

// --- sendShiftedEmails Tests ---

func TestSendShiftedEmails_SendsToAllRecipients(t *testing.T) {
	repo := newMockSessionRepo()
	emailSender := &mockNotificationEmailSender{}
	sessionID := uuid.New()
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			sessionID: {
				{Name: "Charlie", Email: "charlie@example.com"},
			},
		},
	}

	svc := NewSessionService(repo, &mockNotifier{})
	svc.SetEmailNotifier(emailSender, rsvpLister)

	oldSession := &model.SpaceSession{
		ID:        sessionID,
		Title:     "Rescheduled Session",
		Date:      "2025-03-01",
		StartTime: "10:00",
		EndTime:   "14:00",
	}
	newSession := &model.SpaceSession{
		ID:        sessionID,
		Title:     "Rescheduled Session",
		Date:      "2025-03-02",
		StartTime: "11:00",
		EndTime:   "15:00",
	}

	svc.sendShiftedEmails(oldSession, newSession)

	sent := emailSender.getSent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(sent))
	}

	expectedSubject := "Session Rescheduled: Rescheduled Session"
	if sent[0].Subject != expectedSubject {
		t.Errorf("got subject %q, want %q", sent[0].Subject, expectedSubject)
	}

	if !containsStr(sent[0].Body, "Hi Charlie,") {
		t.Error("email should address Charlie")
	}
	if !containsStr(sent[0].Body, "Previously: March 1, 2025, 10:00 – 14:00") {
		t.Errorf("email should contain old time, got: %s", sent[0].Body)
	}
	if !containsStr(sent[0].Body, "Now:        March 2, 2025, 11:00 – 15:00") {
		t.Errorf("email should contain new time, got: %s", sent[0].Body)
	}
}

func TestSendShiftedEmails_NilDependencies(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, &mockNotifier{})

	// Should not panic
	svc.sendShiftedEmails(&model.SpaceSession{}, &model.SpaceSession{})
}

// --- Integration with Cancel/Update flows ---

func TestCancelSession_TriggersEmailNotification(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	emailSender := &mockNotificationEmailSender{}

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled")
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			existing.ID: {
				{Name: "Alice", Email: "alice@example.com"},
			},
		},
	}

	svc := NewSessionService(repo, notifier)
	svc.SetEmailNotifier(emailSender, rsvpLister)

	_, err := svc.Cancel(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give goroutines time to fire
	time.Sleep(50 * time.Millisecond)

	if notifier.canceledCalls != 1 {
		t.Errorf("expected 1 telegram notification, got %d", notifier.canceledCalls)
	}

	sent := emailSender.getSent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 cancel email, got %d", len(sent))
	}
	if sent[0].To != "alice@example.com" {
		t.Errorf("expected email to alice, got %q", sent[0].To)
	}
}

func TestUpdateSession_ShiftTriggersEmailNotification(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	emailSender := &mockNotificationEmailSender{}

	newDate := time.Now().AddDate(0, 0, 14).Format("2006-01-02")
	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled")
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			existing.ID: {
				{Name: "Bob", Email: "bob@example.com"},
			},
		},
	}

	svc := NewSessionService(repo, notifier)
	svc.SetEmailNotifier(emailSender, rsvpLister)

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Date: ptrStr(newDate),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give goroutines time to fire
	time.Sleep(50 * time.Millisecond)

	if notifier.shiftedCalls != 1 {
		t.Errorf("expected 1 shifted notification, got %d", notifier.shiftedCalls)
	}

	sent := emailSender.getSent()
	if len(sent) != 1 {
		t.Fatalf("expected 1 reschedule email, got %d", len(sent))
	}
	if sent[0].To != "bob@example.com" {
		t.Errorf("expected email to bob, got %q", sent[0].To)
	}
	if !containsStr(sent[0].Subject, "Rescheduled") {
		t.Errorf("expected reschedule subject, got %q", sent[0].Subject)
	}
}

func TestUpdateSession_TitleOnlyNoEmail(t *testing.T) {
	repo := newMockSessionRepo()
	emailSender := &mockNotificationEmailSender{}

	existing := repo.addSession("Old Title", futureDate(), "14:00", "18:00", "scheduled")
	rsvpLister := &mockRSVPMemberLister{
		recipients: map[uuid.UUID][]model.RSVPRecipient{
			existing.ID: {
				{Name: "Alice", Email: "alice@example.com"},
			},
		},
	}

	svc := NewSessionService(repo, &mockNotifier{})
	svc.SetEmailNotifier(emailSender, rsvpLister)

	_, err := svc.Update(context.Background(), existing.ID, model.UpdateSessionRequest{
		Title: ptrStr("New Title"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Give goroutines time
	time.Sleep(50 * time.Millisecond)

	sent := emailSender.getSent()
	if len(sent) != 0 {
		t.Errorf("expected 0 emails for title-only update, got %d", len(sent))
	}
}

func TestCancelSession_WithoutEmailNotifier_StillWorks(t *testing.T) {
	repo := newMockSessionRepo()
	notifier := &mockNotifier{}
	svc := NewSessionService(repo, notifier)
	// No SetEmailNotifier — should still work fine

	existing := repo.addSession("Session", futureDate(), "14:00", "18:00", "scheduled")

	session, err := svc.Cancel(context.Background(), existing.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.Status != "canceled" {
		t.Errorf("expected canceled status, got %q", session.Status)
	}

	time.Sleep(50 * time.Millisecond)
	if notifier.canceledCalls != 1 {
		t.Errorf("expected 1 telegram notification, got %d", notifier.canceledCalls)
	}
}

// --- Helper ---

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
