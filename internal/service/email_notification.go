package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/developer-space/api/internal/model"
)

// NotificationEmailSender sends notification emails (fire-and-forget).
type NotificationEmailSender interface {
	SendNotificationEmail(ctx context.Context, to, subject, body string)
}

// RSVPMemberLister retrieves RSVPed member info for email notifications.
type RSVPMemberLister interface {
	ListEmailsBySession(ctx context.Context, sessionID uuid.UUID) ([]model.RSVPRecipient, error)
}


// SetEmailNotifier configures the session service to send email notifications
// on session cancel/reschedule. Call after NewSessionService.
func (s *SessionService) SetEmailNotifier(emailSender NotificationEmailSender, rsvpLister RSVPMemberLister) {
	s.emailSender = emailSender
	s.rsvpLister = rsvpLister
}

// sendCancelEmails sends cancellation emails to all RSVPed members.
// Runs as fire-and-forget; failures are logged but never returned.
func (s *SessionService) sendCancelEmails(session *model.SpaceSession) {
	if s.emailSender == nil || s.rsvpLister == nil {
		return
	}

	ctx := context.Background()
	recipients, err := s.rsvpLister.ListEmailsBySession(ctx, session.ID)
	if err != nil {
		slog.Warn("failed to list RSVP recipients for cancel email", "session_id", session.ID, "error", err)
		return
	}

	subject := cancelEmailSubject(session.Title)
	for _, r := range recipients {
		body := cancelEmailBody(r.Name, session)
		s.emailSender.SendNotificationEmail(ctx, r.Email, subject, body)
	}
}

// sendShiftedEmails sends reschedule emails to all RSVPed members.
// oldSession contains the pre-update values for comparison.
// Runs as fire-and-forget; failures are logged but never returned.
func (s *SessionService) sendShiftedEmails(oldSession, newSession *model.SpaceSession) {
	if s.emailSender == nil || s.rsvpLister == nil {
		return
	}

	ctx := context.Background()
	recipients, err := s.rsvpLister.ListEmailsBySession(ctx, newSession.ID)
	if err != nil {
		slog.Warn("failed to list RSVP recipients for reschedule email", "session_id", newSession.ID, "error", err)
		return
	}

	subject := rescheduleEmailSubject(newSession.Title)
	for _, r := range recipients {
		body := rescheduleEmailBody(r.Name, oldSession, newSession)
		s.emailSender.SendNotificationEmail(ctx, r.Email, subject, body)
	}
}
