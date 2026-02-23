package models

import (
	"time"

	"github.com/google/uuid"
)

type Manifesto struct {
	ID            uuid.UUID  `json:"id"`
	PoliticianID  uuid.UUID  `json:"politician_id"`
	ElectionID    *uuid.UUID `json:"election_id,omitempty"`
	Title         string     `json:"title"`
	Summary       *string    `json:"summary,omitempty"`
	DocumentURL   *string    `json:"document_url,omitempty"`
	PublishedDate *time.Time `json:"published_date,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type PolicyPosition struct {
	ID          uuid.UUID `json:"id"`
	ManifestoID uuid.UUID `json:"manifesto_id"`
	Sector      string    `json:"sector"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	SourceURL   *string   `json:"source_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
