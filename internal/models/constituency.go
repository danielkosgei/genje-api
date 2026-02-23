package models

import (
	"time"

	"github.com/google/uuid"
)

type County struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type Constituency struct {
	ID               uuid.UUID `json:"id"`
	CountyID         uuid.UUID `json:"county_id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	RegisteredVoters int       `json:"registered_voters"`
	CreatedAt        time.Time `json:"created_at"`
}

type Ward struct {
	ID               uuid.UUID `json:"id"`
	ConstituencyID   uuid.UUID `json:"constituency_id"`
	Code             string    `json:"code"`
	Name             string    `json:"name"`
	Slug             string    `json:"slug"`
	RegisteredVoters int       `json:"registered_voters"`
	CreatedAt        time.Time `json:"created_at"`
}
