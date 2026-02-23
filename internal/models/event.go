package models

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	EventType   string     `json:"event_type"`
	Location    *string    `json:"location,omitempty"`
	Latitude    *float64   `json:"latitude,omitempty"`
	Longitude   *float64   `json:"longitude,omitempty"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	SourceURL   *string    `json:"source_url,omitempty"`
	SourceID    *uuid.UUID `json:"source_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type EventParticipant struct {
	EventID      uuid.UUID `json:"event_id"`
	PoliticianID uuid.UUID `json:"politician_id"`
	Role         *string   `json:"role,omitempty"`
}

type EventDetail struct {
	Event
	Participants []PoliticianSummary `json:"participants"`
}
