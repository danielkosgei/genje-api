package models

import (
	"time"

	"github.com/google/uuid"
)

type SentimentSnapshot struct {
	ID           uuid.UUID `json:"id"`
	PoliticianID uuid.UUID `json:"politician_id"`
	Date         time.Time `json:"date"`
	Platform     string    `json:"platform"`
	Score        *float64  `json:"score,omitempty"`
	SampleSize   int       `json:"sample_size"`
	PositivePct  *float64  `json:"positive_pct,omitempty"`
	NegativePct  *float64  `json:"negative_pct,omitempty"`
	NeutralPct   *float64  `json:"neutral_pct,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
