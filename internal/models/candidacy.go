package models

import (
	"time"

	"github.com/google/uuid"
)

type Candidacy struct {
	ID              uuid.UUID  `json:"id"`
	PoliticianID    uuid.UUID  `json:"politician_id"`
	ElectionID      uuid.UUID  `json:"election_id"`
	PositionID      uuid.UUID  `json:"position_id"`
	PartyID         *uuid.UUID `json:"party_id,omitempty"`
	Status          string     `json:"status"`
	DeclarationDate *time.Time `json:"declaration_date,omitempty"`
	ClearanceDate   *time.Time `json:"clearance_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CandidacyDetail struct {
	Candidacy
	PoliticianName string  `json:"politician_name"`
	PoliticianSlug string  `json:"politician_slug"`
	PartyName      *string `json:"party_name,omitempty"`
	PositionTitle  string  `json:"position_title"`
	ElectionName   string  `json:"election_name"`
}
