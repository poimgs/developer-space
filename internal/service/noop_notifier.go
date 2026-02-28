package service

import (
	"log/slog"

	"github.com/developer-space/api/internal/model"
)

// NoopNotifier is a no-op implementation of the Notifier interface.
type NoopNotifier struct{}

func (n *NoopNotifier) SessionCreated(session *model.SpaceSession) {
	slog.Debug("notification skipped (noop)", "event", "session_created", "session_id", session.ID)
}

func (n *NoopNotifier) SessionsCreatedRecurring(sessions []model.SpaceSession) {
	slog.Debug("notification skipped (noop)", "event", "sessions_created_recurring", "count", len(sessions))
}

func (n *NoopNotifier) SessionShifted(session *model.SpaceSession) {
	slog.Debug("notification skipped (noop)", "event", "session_shifted", "session_id", session.ID)
}

func (n *NoopNotifier) SessionCanceled(session *model.SpaceSession) {
	slog.Debug("notification skipped (noop)", "event", "session_canceled", "session_id", session.ID)
}

func (n *NoopNotifier) MemberRSVPed(session *model.SpaceSession, member *model.Member) {
	slog.Debug("notification skipped (noop)", "event", "member_rsvped", "session_id", session.ID)
}

func (n *NoopNotifier) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {
	slog.Debug("notification skipped (noop)", "event", "member_canceled_rsvp", "session_id", session.ID)
}
