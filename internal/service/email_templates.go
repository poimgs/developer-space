package service

import (
	"fmt"
	"time"

	"github.com/developer-space/api/internal/model"
)

func cancelEmailSubject(sessionTitle string) string {
	return fmt.Sprintf("Session Canceled: %s", sessionTitle)
}

func cancelEmailBody(memberName string, session *model.SpaceSession) string {
	return fmt.Sprintf(
		"Hi %s,\n\nThe session \"%s\" scheduled for %s\n(%s – %s) has been canceled.\n\nYour RSVP has been noted. No further action is needed.\n\n— Co-Working Space\n",
		memberName,
		session.Title,
		formatDateHuman(session.Date),
		session.StartTime,
		session.EndTime,
	)
}

func rescheduleEmailSubject(sessionTitle string) string {
	return fmt.Sprintf("Session Rescheduled: %s", sessionTitle)
}

func rescheduleEmailBody(memberName string, oldSession, newSession *model.SpaceSession) string {
	return fmt.Sprintf(
		"Hi %s,\n\nThe session \"%s\" has been rescheduled:\n\n  Previously: %s, %s – %s\n  Now:        %s, %s – %s\n\nYour RSVP is still active — no action needed unless the new time doesn't work for you.\n\n— Co-Working Space\n",
		memberName,
		newSession.Title,
		formatDateHuman(oldSession.Date), oldSession.StartTime, oldSession.EndTime,
		formatDateHuman(newSession.Date), newSession.StartTime, newSession.EndTime,
	)
}

func formatDateHuman(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}
