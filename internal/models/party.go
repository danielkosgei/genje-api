package models

import (
	"time"

	"github.com/google/uuid"
)

type PoliticalParty struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	Abbreviation *string    `json:"abbreviation,omitempty"`
	Slug         string     `json:"slug"`
	LogoURL      *string    `json:"logo_url,omitempty"`
	FoundedDate  *time.Time `json:"founded_date,omitempty"`
	LeaderID     *uuid.UUID `json:"leader_id,omitempty"`
	Ideology     *string    `json:"ideology,omitempty"`
	Website      *string    `json:"website,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type PartyMembership struct {
	ID           uuid.UUID  `json:"id"`
	PoliticianID uuid.UUID  `json:"politician_id"`
	PartyID      uuid.UUID  `json:"party_id"`
	PartyName    string     `json:"party_name,omitempty"`
	JoinedDate   *time.Time `json:"joined_date,omitempty"`
	LeftDate     *time.Time `json:"left_date,omitempty"`
	Role         *string    `json:"role,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PartyWithLeader struct {
	PoliticalParty
	Leader *PoliticianSummary `json:"leader,omitempty"`
}
