package model

import (
	"time"

	"github.com/google/uuid"
)

type SpaceSession struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Date        string     `json:"date"`
	StartTime   string     `json:"start_time"`
	EndTime     string     `json:"end_time"`
	Capacity    int        `json:"capacity"`
	Status      string     `json:"status"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RSVPCount   int        `json:"rsvp_count"`
	UserRSVPed  bool       `json:"user_rsvped,omitempty"`
}

type CreateSessionRequest struct {
	Title        string  `json:"title"`
	Description  *string `json:"description"`
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	Capacity     int     `json:"capacity"`
	RepeatWeekly int     `json:"repeat_weekly"`
}

type UpdateSessionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Date        *string `json:"date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Capacity    *int    `json:"capacity"`
}
