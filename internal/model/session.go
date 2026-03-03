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
	ImageURL    *string    `json:"image_url"`
	Location    *string    `json:"location"`
	SeriesID    *uuid.UUID `json:"series_id"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RSVPCount   int        `json:"rsvp_count"`
	UserRSVPed  bool       `json:"user_rsvped,omitempty"`
}

type CreateSessionRequest struct {
	Title         string     `json:"title"`
	Description   *string    `json:"description"`
	Date          string     `json:"date"`
	StartTime     string     `json:"start_time"`
	EndTime       string     `json:"end_time"`
	Capacity      int        `json:"capacity"`
	Location      *string    `json:"location"`
	RepeatWeekly  int        `json:"repeat_weekly"`
	RepeatForever bool       `json:"repeat_forever"`
	SeriesID      *uuid.UUID `json:"-"`
}

type UpdateSessionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Date        *string `json:"date"`
	StartTime   *string `json:"start_time"`
	EndTime     *string `json:"end_time"`
	Capacity    *int    `json:"capacity"`
	Location    *string `json:"location"`
}
