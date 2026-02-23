package models

import (
	"time"

	"github.com/google/uuid"
)

type ElectionResult struct {
	ID               uuid.UUID  `json:"id"`
	CandidacyID      uuid.UUID  `json:"candidacy_id"`
	PollingStationID *uuid.UUID `json:"polling_station_id,omitempty"`
	Votes            int        `json:"votes"`
	IsFinal          bool       `json:"is_final"`
	Level            string     `json:"level"`
	ReportedAt       time.Time  `json:"reported_at"`
}

type ResultSummary struct {
	CandidacyID    uuid.UUID `json:"candidacy_id"`
	PoliticianName string    `json:"politician_name"`
	PartyName      *string   `json:"party_name,omitempty"`
	TotalVotes     int       `json:"total_votes"`
	Percentage     float64   `json:"percentage"`
	IsFinal        bool      `json:"is_final"`
}
