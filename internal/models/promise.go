package models

import (
	"time"

	"github.com/google/uuid"
)

type Promise struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	Description  string     `json:"description"`
	Sector       *string    `json:"sector,omitempty"`
	MadeDate     *time.Time `json:"made_date,omitempty"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	Status       string     `json:"status"`
	Evidence     *string    `json:"evidence,omitempty"`
	SourceURL    *string    `json:"source_url,omitempty"`
	SourceID     *uuid.UUID `json:"source_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PromiseStats struct {
	Total              int     `json:"total"`
	Fulfilled          int     `json:"fulfilled"`
	Broken             int     `json:"broken"`
	InProgress         int     `json:"in_progress"`
	Pending            int     `json:"pending"`
	PartiallyFulfilled int     `json:"partially_fulfilled"`
	FulfillmentRate    float64 `json:"fulfillment_rate"`
}
