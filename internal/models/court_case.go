package models

import (
	"time"

	"github.com/google/uuid"
)

type CourtCase struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	CaseNumber   *string    `json:"case_number,omitempty"`
	CourtName    *string    `json:"court_name,omitempty"`
	CaseType     string     `json:"case_type"`
	Title        string     `json:"title"`
	Description  *string    `json:"description,omitempty"`
	FilingDate   *time.Time `json:"filing_date,omitempty"`
	Status       string     `json:"status"`
	Outcome      *string    `json:"outcome,omitempty"`
	SourceURL    *string    `json:"source_url,omitempty"`
	SourceID     *uuid.UUID `json:"source_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type IntegrityFlag struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	FlagType     string     `json:"flag_type"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	SourceURL    *string    `json:"source_url,omitempty"`
	SourceID     *uuid.UUID `json:"source_id,omitempty"`
	FlaggedAt    time.Time  `json:"flagged_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
