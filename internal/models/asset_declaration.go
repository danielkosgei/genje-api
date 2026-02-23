package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AssetDeclaration struct {
	ID               uuid.UUID       `json:"id"`
	PoliticianID     uuid.UUID       `json:"politician_id"`
	DeclarationYear  int             `json:"declaration_year"`
	TotalAssets      *float64        `json:"total_assets,omitempty"`
	TotalLiabilities *float64        `json:"total_liabilities,omitempty"`
	Details          json.RawMessage `json:"details"`
	SourceURL        *string         `json:"source_url,omitempty"`
	SourceID         *uuid.UUID      `json:"source_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}
