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

// sendSeriesCancelEmails sends cancellation emails to all RSVPed members across the canceled sessions.
// Deduplicates recipients so each member gets one email listing all their canceled dates.
func (s *SessionService) sendSeriesCancelEmails(series *model.SessionSeries, canceled []model.SpaceSession) {
	if s.emailSender == nil || s.rsvpLister == nil {
		return
	}

	ctx := context.Background()
	recipientDates := map[string][]string{}  // email -> dates
	recipientNames := map[string]string{}    // email -> name

	for _, session := range canceled {
		recipients, err := s.rsvpLister.ListEmailsBySession(ctx, session.ID)
		if err != nil {
			slog.Warn("failed to list RSVP recipients for series cancel email", "session_id", session.ID, "error", err)
			continue
		}
		for _, r := range recipients {
			recipientDates[r.Email] = append(recipientDates[r.Email], session.Date)
			recipientNames[r.Email] = r.Name
		}
	}

	subject := seriesCanceledEmailSubject(series.Title)
	for email, dates := range recipientDates {
		body := seriesCanceledEmailBody(recipientNames[email], series, dates)
		s.emailSender.SendNotificationEmail(ctx, email, subject, body)
	}
}

// sendSeriesUpdatedEmails sends update emails to all RSVPed members across the affected sessions.
// Deduplicates recipients so each member gets one email listing all their affected dates.
func (s *SessionService) sendSeriesUpdatedEmails(series *model.SessionSeries, affected []model.SpaceSession) {
	if s.emailSender == nil || s.rsvpLister == nil {
		return
	}

	ctx := context.Background()
	recipientDates := map[string][]string{}
	recipientNames := map[string]string{}

	for _, session := range affected {
		recipients, err := s.rsvpLister.ListEmailsBySession(ctx, session.ID)
		if err != nil {
			slog.Warn("failed to list RSVP recipients for series update email", "session_id", session.ID, "error", err)
			continue
		}
		for _, r := range recipients {
			recipientDates[r.Email] = append(recipientDates[r.Email], session.Date)
			recipientNames[r.Email] = r.Name
		}
	}

	subject := seriesUpdatedEmailSubject(series.Title)
	for email, dates := range recipientDates {
		body := seriesUpdatedEmailBody(recipientNames[email], series, dates)
		s.emailSender.SendNotificationEmail(ctx, email, subject, body)
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
