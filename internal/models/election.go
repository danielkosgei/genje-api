package models

import (
	"time"

	"github.com/google/uuid"
)

type Election struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	ElectionDate *time.Time `json:"election_date,omitempty"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type ElectivePosition struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	Level           string     `json:"level"`
	CountyID        *uuid.UUID `json:"county_id,omitempty"`
	ConstituencyID  *uuid.UUID `json:"constituency_id,omitempty"`
	WardID          *uuid.UUID `json:"ward_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}
