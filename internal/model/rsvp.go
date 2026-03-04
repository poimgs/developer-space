package model

import (
	"time"

	"github.com/google/uuid"
)

type RSVP struct {
	ID        uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	MemberID  uuid.UUID `json:"member_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RSVPWithMember struct {
	ID        uuid.UUID  `json:"id"`
	SessionID uuid.UUID  `json:"session_id"`
	Member    RSVPMember `json:"member"`
	CreatedAt time.Time  `json:"created_at"`
}

type RSVPMember struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	TelegramHandle *string   `json:"telegram_handle"`
	Bio            *string   `json:"bio"`
}

// RSVPRecipient contains the minimum member info needed for email notifications.
type RSVPRecipient struct {
	Name  string
	Email string
}
