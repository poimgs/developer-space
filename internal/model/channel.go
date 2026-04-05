package model

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Type      string     `json:"type"`
	SessionID *uuid.UUID `json:"session_id"`
	CreatedBy uuid.UUID  `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateChannelRequest struct {
	Name string `json:"name"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	MemberID  uuid.UUID `json:"member_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageWithAuthor struct {
	Message
	AuthorName string `json:"author_name"`
}

type MessagePage struct {
	Messages []MessageWithAuthor `json:"messages"`
	Cursor   *string             `json:"cursor"`
}
