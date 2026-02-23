package models

import (
	"time"

	"github.com/google/uuid"
)

type Source struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	URL            *string    `json:"url,omitempty"`
	Type           string     `json:"type"`
	Reliability    string     `json:"reliability"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
