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
		"Hi %s,\n\nThe session \"%s\" scheduled for %s\n(%s – %s) has been canceled.\n\nYour RSVP has been noted. No further action is needed.\n\n— Developer Space\n",
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
		"Hi %s,\n\nThe session \"%s\" has been rescheduled:\n\n  Previously: %s, %s – %s\n  Now:        %s, %s – %s\n\nYour RSVP is still active — no action needed unless the new time doesn't work for you.\n\n— Developer Space\n",
		memberName,
		newSession.Title,
		formatDateHuman(oldSession.Date), oldSession.StartTime, oldSession.EndTime,
		formatDateHuman(newSession.Date), newSession.StartTime, newSession.EndTime,
	)
}

func seriesUpdatedEmailSubject(seriesTitle string) string {
	return fmt.Sprintf("Recurring Series Updated: %s", seriesTitle)
}

func seriesUpdatedEmailBody(memberName string, series *model.SessionSeries, dates []string) string {
	dateList := ""
	for _, d := range dates {
		dateList += fmt.Sprintf("  - %s\n", formatDateHuman(d))
	}
	return fmt.Sprintf(
		"Hi %s,\n\nThe recurring series \"%s\" has been updated.\n\nAffected sessions:\n%s\nYour RSVPs are still active — no action needed unless the changes don't work for you.\n\n— Developer Space\n",
		memberName,
		series.Title,
		dateList,
	)
}

func seriesCanceledEmailSubject(seriesTitle string) string {
	return fmt.Sprintf("Recurring Series Canceled: %s", seriesTitle)
}

func seriesCanceledEmailBody(memberName string, series *model.SessionSeries, dates []string) string {
	dateList := ""
	for _, d := range dates {
		dateList += fmt.Sprintf("  - %s\n", formatDateHuman(d))
	}
	return fmt.Sprintf(
		"Hi %s,\n\nThe recurring series \"%s\" has been canceled.\n\nCanceled sessions:\n%s\nYour RSVPs have been noted. No further action is needed.\n\n— Developer Space\n",
		memberName,
		series.Title,
		dateList,
	)
}

func formatDateHuman(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("January 2, 2006")
}
