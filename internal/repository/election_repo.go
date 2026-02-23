package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type ElectionRepo struct {
	pool *pgxpool.Pool
}

func NewElectionRepo(pool *pgxpool.Pool) *ElectionRepo {
	return &ElectionRepo{pool: pool}
}

func (r *ElectionRepo) List(ctx context.Context) ([]models.Election, error) {
	query := `
		SELECT id, name, election_date, type, status, created_at, updated_at
		FROM elections
		ORDER BY election_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list elections: %w", err)
	}
	defer rows.Close()

	var elections []models.Election
	for rows.Next() {
		var e models.Election
		if err := rows.Scan(&e.ID, &e.Name, &e.ElectionDate, &e.Type, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan election: %w", err)
		}
		elections = append(elections, e)
	}
	return elections, nil
}

func (r *ElectionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Election, error) {
	query := `
		SELECT id, name, election_date, type, status, created_at, updated_at
		FROM elections WHERE id = $1`

	var e models.Election
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.Name, &e.ElectionDate, &e.Type, &e.Status, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get election: %w", err)
	}
	return &e, nil
}

func (r *ElectionRepo) GetCandidates(ctx context.Context, electionID uuid.UUID) ([]models.CandidacyDetail, error) {
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
		WHERE c.election_id = $1
		ORDER BY ep.title, p.last_name`

	rows, err := r.pool.Query(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("get election candidates: %w", err)
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

func (r *ElectionRepo) GetResults(ctx context.Context, electionID uuid.UUID) ([]models.ResultSummary, error) {
	query := `
		SELECT c.id,
		       p.first_name || ' ' || p.last_name,
		       pp.name,
		       COALESCE(SUM(er.votes), 0),
		       CASE WHEN SUM(SUM(er.votes)) OVER () > 0
		            THEN ROUND(SUM(er.votes)::numeric / SUM(SUM(er.votes)) OVER () * 100, 2)
		            ELSE 0 END,
		       BOOL_AND(er.is_final)
		FROM candidacies c
		JOIN politicians p ON p.id = c.politician_id
		LEFT JOIN political_parties pp ON pp.id = c.party_id
		LEFT JOIN election_results er ON er.candidacy_id = c.id
		WHERE c.election_id = $1
		GROUP BY c.id, p.first_name, p.last_name, pp.name
		ORDER BY COALESCE(SUM(er.votes), 0) DESC`

	rows, err := r.pool.Query(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("get results: %w", err)
	}
	defer rows.Close()

	var results []models.ResultSummary
	for rows.Next() {
		var rs models.ResultSummary
		if err := rows.Scan(&rs.CandidacyID, &rs.PoliticianName, &rs.PartyName, &rs.TotalVotes, &rs.Percentage, &rs.IsFinal); err != nil {
			return nil, fmt.Errorf("scan result: %w", err)
		}
		results = append(results, rs)
	}
	return results, nil
}

func (r *ElectionRepo) GetTimeline(ctx context.Context, electionID uuid.UUID) ([]models.TimelineEvent, error) {
	query := `
		SELECT id, election_id, title, description, milestone_type, date, status, created_at, updated_at
		FROM election_timeline
		WHERE election_id = $1
		ORDER BY date`

	rows, err := r.pool.Query(ctx, query, electionID)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer rows.Close()

	var events []models.TimelineEvent
	for rows.Next() {
		var te models.TimelineEvent
		if err := rows.Scan(&te.ID, &te.ElectionID, &te.Title, &te.Description, &te.MilestoneType, &te.Date, &te.Status, &te.CreatedAt, &te.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan timeline event: %w", err)
		}
		events = append(events, te)
	}
	return events, nil
}
