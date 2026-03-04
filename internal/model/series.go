package model

import (
	"time"

	"github.com/google/uuid"
)

type SessionSeries struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	DayOfWeek   int       `json:"day_of_week"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	Capacity    int       `json:"capacity"`
	EveryNWeeks int       `json:"every_n_weeks"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
