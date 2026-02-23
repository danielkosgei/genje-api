package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type GeographyRepo struct {
	pool *pgxpool.Pool
}

func NewGeographyRepo(pool *pgxpool.Pool) *GeographyRepo {
	return &GeographyRepo{pool: pool}
}

func (r *GeographyRepo) ListCounties(ctx context.Context) ([]models.County, error) {
	query := `SELECT id, code, name, slug, created_at FROM counties ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list counties: %w", err)
	}
	defer rows.Close()

	var counties []models.County
	for rows.Next() {
		var c models.County
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan county: %w", err)
		}
		counties = append(counties, c)
	}
	return counties, nil
}

func (r *GeographyRepo) GetCountyByCode(ctx context.Context, code string) (*models.County, error) {
	query := `SELECT id, code, name, slug, created_at FROM counties WHERE code = $1`

	var c models.County
	err := r.pool.QueryRow(ctx, query, code).Scan(&c.ID, &c.Code, &c.Name, &c.Slug, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get county: %w", err)
	}
	return &c, nil
}

func (r *GeographyRepo) GetConstituenciesByCounty(ctx context.Context, countyCode string) ([]models.Constituency, error) {
	query := `
		SELECT c.id, c.county_id, c.code, c.name, c.slug, c.registered_voters, c.created_at
		FROM constituencies c
		JOIN counties co ON co.id = c.county_id
		WHERE co.code = $1
		ORDER BY c.name`

	rows, err := r.pool.Query(ctx, query, countyCode)
	if err != nil {
		return nil, fmt.Errorf("get constituencies: %w", err)
	}
	defer rows.Close()

	var constituencies []models.Constituency
	for rows.Next() {
		var c models.Constituency
		if err := rows.Scan(&c.ID, &c.CountyID, &c.Code, &c.Name, &c.Slug, &c.RegisteredVoters, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan constituency: %w", err)
		}
		constituencies = append(constituencies, c)
	}
	return constituencies, nil
}

func (r *GeographyRepo) GetConstituencyByCode(ctx context.Context, code string) (*models.Constituency, error) {
	query := `SELECT id, county_id, code, name, slug, registered_voters, created_at FROM constituencies WHERE code = $1`

	var c models.Constituency
	err := r.pool.QueryRow(ctx, query, code).Scan(&c.ID, &c.CountyID, &c.Code, &c.Name, &c.Slug, &c.RegisteredVoters, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get constituency: %w", err)
	}
	return &c, nil
}

func (r *GeographyRepo) GetWardsByConstituency(ctx context.Context, constituencyCode string) ([]models.Ward, error) {
	query := `
		SELECT w.id, w.constituency_id, w.code, w.name, w.slug, w.registered_voters, w.created_at
		FROM wards w
		JOIN constituencies c ON c.id = w.constituency_id
		WHERE c.code = $1
		ORDER BY w.name`

	rows, err := r.pool.Query(ctx, query, constituencyCode)
	if err != nil {
		return nil, fmt.Errorf("get wards: %w", err)
	}
	defer rows.Close()

	var wards []models.Ward
	for rows.Next() {
		var w models.Ward
		if err := rows.Scan(&w.ID, &w.ConstituencyID, &w.Code, &w.Name, &w.Slug, &w.RegisteredVoters, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan ward: %w", err)
		}
		wards = append(wards, w)
	}
	return wards, nil
}

func (r *GeographyRepo) GetPollingStationsByConstituency(ctx context.Context, constituencyCode string) ([]models.PollingStation, error) {
	query := `
		SELECT ps.id, ps.ward_id, ps.code, ps.name, ps.latitude, ps.longitude, ps.registered_voters, ps.created_at
		FROM polling_stations ps
		JOIN wards w ON w.id = ps.ward_id
		JOIN constituencies c ON c.id = w.constituency_id
		WHERE c.code = $1
		ORDER BY ps.name`

	rows, err := r.pool.Query(ctx, query, constituencyCode)
	if err != nil {
		return nil, fmt.Errorf("get polling stations: %w", err)
	}
	defer rows.Close()

	var stations []models.PollingStation
	for rows.Next() {
		var s models.PollingStation
		if err := rows.Scan(&s.ID, &s.WardID, &s.Code, &s.Name, &s.Latitude, &s.Longitude, &s.RegisteredVoters, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan polling station: %w", err)
		}
		stations = append(stations, s)
	}
	return stations, nil
}

func (r *GeographyRepo) GetCandidatesByConstituency(ctx context.Context, constituencyCode string) ([]models.CandidacyDetail, error) {
	query := `
		SELECT c.id, c.politician_id, c.election_id, c.position_id, c.party_id,
		       c.status, c.declaration_date, c.clearance_date, c.created_at, c.updated_at,
		       p.first_name || ' ' || p.last_name, p.slug,
		       pp.name, ep.title, e.name
		FROM candidacies c
		JOIN politicians p ON p.id = c.politician_id
		JOIN elective_positions ep ON ep.id = c.position_id
		JOIN elections e ON e.id = c.election_id
		LEFT JOIN political_parties pp ON pp.id = c.party_id
		WHERE ep.constituency_id = (SELECT id FROM constituencies WHERE code = $1)
		  AND e.status != 'completed'
		ORDER BY ep.title, p.last_name`

	rows, err := r.pool.Query(ctx, query, constituencyCode)
	if err != nil {
		return nil, fmt.Errorf("get constituency candidates: %w", err)
	}
	defer rows.Close()

	var candidates []models.CandidacyDetail
	for rows.Next() {
		var cd models.CandidacyDetail
		if err := rows.Scan(
			&cd.ID, &cd.PoliticianID, &cd.ElectionID, &cd.PositionID, &cd.PartyID,
			&cd.Status, &cd.DeclarationDate, &cd.ClearanceDate, &cd.CreatedAt, &cd.UpdatedAt,
			&cd.PoliticianName, &cd.PoliticianSlug,
			&cd.PartyName, &cd.PositionTitle, &cd.ElectionName,
		); err != nil {
			return nil, fmt.Errorf("scan candidacy: %w", err)
		}
		candidates = append(candidates, cd)
	}
	return candidates, nil
}
