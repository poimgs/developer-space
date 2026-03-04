package telegram

import (
	"fmt"
	"strings"

	"github.com/developer-space/api/internal/model"
)

// TelegramNotifier implements the service.Notifier interface by formatting
// notification messages and dispatching them via TelegramService.
type TelegramNotifier struct {
	svc *TelegramService
}

// NewTelegramNotifier creates a notifier that sends messages via the given TelegramService.
func NewTelegramNotifier(svc *TelegramService) *TelegramNotifier {
	return &TelegramNotifier{svc: svc}
}

// SessionCreated sends a "New Session" notification.
func (n *TelegramNotifier) SessionCreated(session *model.SpaceSession) {
	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString("📅 *New Session*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(session.Title)))
	sb.WriteString(fmt.Sprintf("📆 %s\n", esc(session.Date)))
	sb.WriteString(fmt.Sprintf("🕐 %s – %s\n", esc(session.StartTime), esc(session.EndTime)))
	if session.Description != nil && *session.Description != "" {
		sb.WriteString(fmt.Sprintf("\n\n%s", esc(*session.Description)))
	}

	n.svc.SendMessage(sb.String())
}

// SessionsCreatedRecurring sends a single summary message for recurring sessions.
func (n *TelegramNotifier) SessionsCreatedRecurring(sessions []model.SpaceSession) {
	if len(sessions) == 0 {
		return
	}

	esc := EscapeMarkdownV2
	first := sessions[0]

	var sb strings.Builder
	sb.WriteString("📅 *Recurring Sessions Created*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(first.Title)))
	sb.WriteString(fmt.Sprintf("🕐 %s – %s\n", esc(first.StartTime), esc(first.EndTime)))

	for _, s := range sessions {
		sb.WriteString(fmt.Sprintf("\n📆 %s", esc(s.Date)))
	}

	n.svc.SendMessage(sb.String())
}

// SessionShifted sends a "Session Rescheduled" notification.
// Note: The notifier receives the updated session. Since we don't have the old
// date/time at this layer, we show the current (new) date and time.
func (n *TelegramNotifier) SessionShifted(session *model.SpaceSession) {
	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString("🔄 *Session Rescheduled*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(session.Title)))
	sb.WriteString(fmt.Sprintf("📆 %s\n", esc(session.Date)))
	sb.WriteString(fmt.Sprintf("🕐 %s – %s", esc(session.StartTime), esc(session.EndTime)))

	n.svc.SendMessage(sb.String())
}

// SessionCanceled sends a "Session Canceled" notification.
func (n *TelegramNotifier) SessionCanceled(session *model.SpaceSession) {
	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString("❌ *Session Canceled*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(session.Title)))
	sb.WriteString(fmt.Sprintf("📆 %s, %s – %s", esc(session.Date), esc(session.StartTime), esc(session.EndTime)))

	n.svc.SendMessage(sb.String())
}

// MemberRSVPed sends an "RSVP" notification.
func (n *TelegramNotifier) MemberRSVPed(session *model.SpaceSession, member *model.Member) {
	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✅ %s RSVPed\n\n", esc(member.Name)))
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(session.Title)))
	sb.WriteString(fmt.Sprintf("📆 %s, %s – %s\n", esc(session.Date), esc(session.StartTime), esc(session.EndTime)))
	sb.WriteString(fmt.Sprintf("👥 %s attending", esc(fmt.Sprintf("%d", session.RSVPCount))))

	n.svc.SendMessage(sb.String())
}

// MemberCanceledRSVP sends an "RSVP Canceled" notification.
func (n *TelegramNotifier) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {
	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🚫 %s canceled RSVP\n\n", esc(member.Name)))
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(session.Title)))
	sb.WriteString(fmt.Sprintf("📆 %s, %s – %s\n", esc(session.Date), esc(session.StartTime), esc(session.EndTime)))
	sb.WriteString(fmt.Sprintf("👥 %s attending", esc(fmt.Sprintf("%d", session.RSVPCount))))

	n.svc.SendMessage(sb.String())
}

// SeriesUpdated sends a "Recurring Series Updated" notification.
func (n *TelegramNotifier) SeriesUpdated(series *model.SessionSeries, affected []model.SpaceSession) {
	if len(affected) == 0 {
		return
	}

	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString("🔄 *Recurring Series Updated*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(series.Title)))
	sb.WriteString(fmt.Sprintf("🕐 %s – %s\n", esc(series.StartTime), esc(series.EndTime)))
	sb.WriteString(fmt.Sprintf("📅 %s sessions affected", esc(fmt.Sprintf("%d", len(affected)))))

	n.svc.SendMessage(sb.String())
}

// SeriesCanceled sends a "Recurring Series Canceled" notification.
func (n *TelegramNotifier) SeriesCanceled(series *model.SessionSeries, canceled []model.SpaceSession) {
	if len(canceled) == 0 {
		return
	}

	esc := EscapeMarkdownV2

	var sb strings.Builder
	sb.WriteString("❌ *Recurring Series Canceled*\n\n")
	sb.WriteString(fmt.Sprintf("*%s*\n", esc(series.Title)))

	for _, s := range canceled {
		sb.WriteString(fmt.Sprintf("\n📆 %s", esc(s.Date)))
	}

	n.svc.SendMessage(sb.String())
}
