package model

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	Name           string     `json:"name"`
	TelegramHandle *string    `json:"telegram_handle"`
	IsAdmin        bool       `json:"is_admin"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateMemberRequest struct {
	Email          string  `json:"email"`
	Name           string  `json:"name"`
	TelegramHandle *string `json:"telegram_handle"`
	IsAdmin        bool    `json:"is_admin"`
	SendInvite     bool    `json:"send_invite"`
}

type UpdateMemberRequest struct {
	Name           *string `json:"name"`
	TelegramHandle *string `json:"telegram_handle"`
	IsAdmin        *bool   `json:"is_admin"`
	IsActive       *bool   `json:"is_active"`
}
