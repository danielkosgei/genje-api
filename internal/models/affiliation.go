package models

import (
	"time"

	"github.com/google/uuid"
)

type Affiliation struct {
	ID                  uuid.UUID  `json:"id"`
	PoliticianID        uuid.UUID  `json:"politician_id"`
	RelatedPoliticianID uuid.UUID  `json:"related_politician_id"`
	RelationshipType    string     `json:"relationship_type"`
	Description         *string    `json:"description,omitempty"`
	SourceURL           *string    `json:"source_url,omitempty"`
	SourceID            *uuid.UUID `json:"source_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
}

type AffiliationDetail struct {
	Affiliation
	RelatedPoliticianName string  `json:"related_politician_name"`
	RelatedPoliticianSlug string  `json:"related_politician_slug"`
	RelatedPhotoURL       *string `json:"related_photo_url,omitempty"`
}
