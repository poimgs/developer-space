package service

import "github.com/developer-space/api/internal/model"

// Notifier defines the interface for sending notifications about session and RSVP events.
type Notifier interface {
	SessionCreated(session *model.SpaceSession)
	SessionsCreatedRecurring(sessions []model.SpaceSession)
	SessionShifted(session *model.SpaceSession)
	SessionCanceled(session *model.SpaceSession)
	MemberRSVPed(session *model.SpaceSession, member *model.Member)
	MemberCanceledRSVP(session *model.SpaceSession, member *model.Member)
}
