package models

import (
	"time"

	"github.com/google/uuid"
)

type Coalition struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	Slug             string     `json:"slug"`
	FormedDate       *time.Time `json:"formed_date,omitempty"`
	DissolvedDate    *time.Time `json:"dissolved_date,omitempty"`
	PrincipalPartyID *uuid.UUID `json:"principal_party_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CoalitionMember struct {
	CoalitionID uuid.UUID  `json:"coalition_id"`
	PartyID     uuid.UUID  `json:"party_id"`
	PartyName   string     `json:"party_name,omitempty"`
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
	LeftAt      *time.Time `json:"left_at,omitempty"`
}

type CoalitionDetail struct {
	Coalition
	PrincipalParty *PoliticalParty   `json:"principal_party,omitempty"`
	Members        []CoalitionMember `json:"members"`
}
