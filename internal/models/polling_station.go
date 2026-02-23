package models

import (
	"time"

	"github.com/google/uuid"
)

type PollingStation struct {
	ID               uuid.UUID `json:"id"`
	WardID           uuid.UUID `json:"ward_id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Latitude         *float64  `json:"latitude,omitempty"`
	Longitude        *float64  `json:"longitude,omitempty"`
	RegisteredVoters int       `json:"registered_voters"`
	CreatedAt        time.Time `json:"created_at"`
}
