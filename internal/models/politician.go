package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Politician struct {
	ID            uuid.UUID       `json:"id"`
	Slug          string          `json:"slug"`
	FirstName     string          `json:"first_name"`
	LastName      string          `json:"last_name"`
	OtherNames    *string         `json:"other_names,omitempty"`
	DateOfBirth   *time.Time      `json:"date_of_birth,omitempty"`
	DateOfDeath   *time.Time      `json:"date_of_death,omitempty"`
	Gender        *string         `json:"gender,omitempty"`
	Status        string          `json:"status"`
	Bio           *string         `json:"bio,omitempty"`
	PhotoURL      *string         `json:"photo_url,omitempty"`
	Education     json.RawMessage `json:"education"`
	CareerHistory json.RawMessage `json:"career_history"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type PoliticianSummary struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Status    string    `json:"status"`
	PhotoURL  *string   `json:"photo_url,omitempty"`
	Party     *string   `json:"current_party,omitempty"`
}

type PoliticianDossier struct {
	Politician
	CurrentParty    *PartyMembership   `json:"current_party,omitempty"`
	PartyHistory    []PartyMembership  `json:"party_history"`
	Candidacies     []CandidacyDetail  `json:"candidacies"`
	IntegrityFlags  []IntegrityFlag    `json:"integrity_flags"`
}

type PoliticianFilter struct {
	Query    string
	PartyID  *uuid.UUID
	CountyID *uuid.UUID
	Position *string
	Limit    int
	Offset   int
}
