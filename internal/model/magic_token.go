package model

import (
	"time"

	"github.com/google/uuid"
)

type MagicToken struct {
	ID        uuid.UUID  `json:"id"`
	MemberID  uuid.UUID  `json:"member_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}
