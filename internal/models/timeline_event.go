package models

import (
	"time"

	"github.com/google/uuid"
)

type TimelineEvent struct {
	ID            uuid.UUID `json:"id"`
	ElectionID    uuid.UUID `json:"election_id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	MilestoneType string    `json:"milestone_type"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
