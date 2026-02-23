package models

import (
	"time"

	"github.com/google/uuid"
)

type Controversy struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	Category     *string    `json:"category,omitempty"`
	Date         *time.Time `json:"date,omitempty"`
	Severity     string     `json:"severity"`
	SourceURL    *string    `json:"source_url,omitempty"`
	SourceID     *uuid.UUID `json:"source_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
