package model

import (
	"time"

	"github.com/google/uuid"
)

type Member struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	TelegramHandle  *string   `json:"telegram_handle"`
	IsAdmin         bool      `json:"is_admin"`
	IsActive        bool      `json:"is_active"`
	Bio             *string   `json:"bio"`
	Skills          []string  `json:"skills"`
	LinkedinURL     *string   `json:"linkedin_url"`
	InstagramHandle *string   `json:"instagram_handle"`
	GithubUsername  *string   `json:"github_username"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateMemberRequest struct {
	Email          string  `json:"email"`
	Name           string  `json:"name"`
	TelegramHandle *string `json:"telegram_handle"`
	IsAdmin        bool    `json:"is_admin"`
	SendInvite     bool    `json:"send_invite"`
}

type UpdateMemberRequest struct {
	Name            *string  `json:"name"`
	TelegramHandle  *string  `json:"telegram_handle"`
	IsAdmin         *bool    `json:"is_admin"`
	IsActive        *bool    `json:"is_active"`
	Bio             *string  `json:"bio"`
	Skills          []string `json:"skills"`
	LinkedinURL     *string  `json:"linkedin_url"`
	InstagramHandle *string  `json:"instagram_handle"`
	GithubUsername  *string  `json:"github_username"`
}

// PublicMember contains only user-controlled public information.
type PublicMember struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	TelegramHandle  *string   `json:"telegram_handle"`
	Bio             *string   `json:"bio"`
	Skills          []string  `json:"skills"`
	LinkedinURL     *string   `json:"linkedin_url"`
	InstagramHandle *string   `json:"instagram_handle"`
	GithubUsername  *string   `json:"github_username"`
}
