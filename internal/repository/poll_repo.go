package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

// PollRepo reads published polls and their results.
type PollRepo struct {
	pool *pgxpool.Pool
}

func NewPollRepo(pool *pgxpool.Pool) *PollRepo {
	return &PollRepo{pool: pool}
}

// List returns one page of polls, newest first, with their results. race "" means every race.
func (r *PollRepo) List(ctx context.Context, race string, limit, offset int) ([]models.Poll, int, error) {
	where := " WHERE ($1 = '' OR p.race = $1)"

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM polls p`+where, race).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count polls: %w", err)
	}

	query := `
		SELECT p.id::text, ps.name, p.race,
		       to_char(p.fieldwork_start, 'YYYY-MM-DD'), to_char(p.fieldwork_end, 'YYYY-MM-DD'),
		       p.sample_size, p.margin_of_error::float8, to_char(p.published_on, 'YYYY-MM-DD'), p.source_url
		FROM polls p
		JOIN pollsters ps ON ps.id = p.pollster_id` + where + `
		ORDER BY p.published_on DESC, p.id
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, race, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list polls: %w", err)
	}
	defer rows.Close()

	var polls []models.Poll
	index := map[string]int{}
	for rows.Next() {
		var p models.Poll
		if err := rows.Scan(&p.ID, &p.Pollster, &p.Race, &p.FieldworkStart, &p.FieldworkEnd,
			&p.SampleSize, &p.MarginOfError, &p.PublishedOn, &p.SourceURL); err != nil {
			return nil, 0, fmt.Errorf("scan poll: %w", err)
		}
		p.Results = []models.PollResult{}
		index[p.ID] = len(polls)
		polls = append(polls, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(polls) == 0 {
		return polls, total, nil
	}

	ids := make([]string, 0, len(polls))
	for _, p := range polls {
		ids = append(ids, p.ID)
	}
	resultRows, err := r.pool.Query(ctx, `
		SELECT poll_id::text, candidate_name, party, share::float8
		FROM poll_results
		WHERE poll_id::text = ANY($1)
		ORDER BY share DESC, candidate_name`, ids)
	if err != nil {
		return nil, 0, fmt.Errorf("list poll results: %w", err)
	}
	defer resultRows.Close()

	for resultRows.Next() {
		var pollID string
		var res models.PollResult
		if err := resultRows.Scan(&pollID, &res.CandidateName, &res.Party, &res.Share); err != nil {
			return nil, 0, fmt.Errorf("scan poll result: %w", err)
		}
		polls[index[pollID]].Results = append(polls[index[pollID]].Results, res)
	}
	return polls, total, resultRows.Err()
}
