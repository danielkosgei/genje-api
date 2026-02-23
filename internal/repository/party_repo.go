package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type PartyRepo struct {
	pool *pgxpool.Pool
}

func NewPartyRepo(pool *pgxpool.Pool) *PartyRepo {
	return &PartyRepo{pool: pool}
}

func (r *PartyRepo) List(ctx context.Context) ([]models.PoliticalParty, error) {
	query := `
		SELECT id, name, abbreviation, slug, logo_url, founded_date, leader_id,
		       ideology, website, status, created_at, updated_at
		FROM political_parties
		ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list parties: %w", err)
	}
	defer rows.Close()

	var parties []models.PoliticalParty
	for rows.Next() {
		var p models.PoliticalParty
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Abbreviation, &p.Slug, &p.LogoURL, &p.FoundedDate,
			&p.LeaderID, &p.Ideology, &p.Website, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan party: %w", err)
		}
		parties = append(parties, p)
	}
	return parties, nil
}

func (r *PartyRepo) GetBySlug(ctx context.Context, slug string) (*models.PartyWithLeader, error) {
	query := `
		SELECT pp.id, pp.name, pp.abbreviation, pp.slug, pp.logo_url, pp.founded_date,
		       pp.leader_id, pp.ideology, pp.website, pp.status, pp.created_at, pp.updated_at,
		       p.id, p.slug, p.first_name, p.last_name, p.photo_url
		FROM political_parties pp
		LEFT JOIN politicians p ON p.id = pp.leader_id
		WHERE pp.slug = $1`

	var party models.PartyWithLeader
	var lSlug, lFirst, lLast *string
	var lPhoto *string
	var lUUID interface{}

	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&party.ID, &party.Name, &party.Abbreviation, &party.Slug, &party.LogoURL,
		&party.FoundedDate, &party.LeaderID, &party.Ideology, &party.Website,
		&party.Status, &party.CreatedAt, &party.UpdatedAt,
		&lUUID, &lSlug, &lFirst, &lLast, &lPhoto,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get party by slug: %w", err)
	}

	if lSlug != nil {
		party.Leader = &models.PoliticianSummary{
			Slug:      *lSlug,
			FirstName: *lFirst,
			LastName:  *lLast,
			PhotoURL:  lPhoto,
		}
	}

	return &party, nil
}

func (r *PartyRepo) GetMembers(ctx context.Context, partySlug string) ([]models.PoliticianSummary, error) {
	query := `
		SELECT p.id, p.slug, p.first_name, p.last_name, p.photo_url
		FROM politicians p
		JOIN party_memberships pm ON pm.politician_id = p.id
		JOIN political_parties pp ON pp.id = pm.party_id
		WHERE pp.slug = $1 AND pm.left_date IS NULL
		ORDER BY p.last_name, p.first_name`

	rows, err := r.pool.Query(ctx, query, partySlug)
	if err != nil {
		return nil, fmt.Errorf("get party members: %w", err)
	}
	defer rows.Close()

	var members []models.PoliticianSummary
	for rows.Next() {
		var m models.PoliticianSummary
		if err := rows.Scan(&m.ID, &m.Slug, &m.FirstName, &m.LastName, &m.PhotoURL); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, m)
	}
	return members, nil
}

func (r *PartyRepo) ListCoalitions(ctx context.Context) ([]models.Coalition, error) {
	query := `
		SELECT id, name, slug, formed_date, dissolved_date, principal_party_id, created_at, updated_at
		FROM coalitions
		ORDER BY formed_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list coalitions: %w", err)
	}
	defer rows.Close()

	var coalitions []models.Coalition
	for rows.Next() {
		var c models.Coalition
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.FormedDate, &c.DissolvedDate, &c.PrincipalPartyID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan coalition: %w", err)
		}
		coalitions = append(coalitions, c)
	}
	return coalitions, nil
}

func (r *PartyRepo) GetCoalitionBySlug(ctx context.Context, slug string) (*models.CoalitionDetail, error) {
	query := `
		SELECT id, name, slug, formed_date, dissolved_date, principal_party_id, created_at, updated_at
		FROM coalitions WHERE slug = $1`

	var cd models.CoalitionDetail
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&cd.ID, &cd.Name, &cd.Slug, &cd.FormedDate, &cd.DissolvedDate,
		&cd.PrincipalPartyID, &cd.CreatedAt, &cd.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get coalition by slug: %w", err)
	}

	membersQuery := `
		SELECT cm.coalition_id, cm.party_id, pp.name, cm.joined_at, cm.left_at
		FROM coalition_members cm
		JOIN political_parties pp ON pp.id = cm.party_id
		WHERE cm.coalition_id = $1
		ORDER BY pp.name`

	rows, err := r.pool.Query(ctx, membersQuery, cd.ID)
	if err != nil {
		return nil, fmt.Errorf("get coalition members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m models.CoalitionMember
		if err := rows.Scan(&m.CoalitionID, &m.PartyID, &m.PartyName, &m.JoinedAt, &m.LeftAt); err != nil {
			return nil, fmt.Errorf("scan coalition member: %w", err)
		}
		cd.Members = append(cd.Members, m)
	}

	return &cd, nil
}
