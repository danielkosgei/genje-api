package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type SentimentRepo struct {
	pool *pgxpool.Pool
}

func NewSentimentRepo(pool *pgxpool.Pool) *SentimentRepo {
	return &SentimentRepo{pool: pool}
}

func (r *SentimentRepo) GetByPolitician(ctx context.Context, politicianID uuid.UUID, limit int) ([]models.SentimentSnapshot, error) {
	if limit <= 0 {
		limit = 30
	}

	query := `
		SELECT id, politician_id, date, platform, score, sample_size,
		       positive_pct, negative_pct, neutral_pct, created_at
		FROM sentiment_snapshots
		WHERE politician_id = $1
		ORDER BY date DESC
		LIMIT $2`

	rows, err := r.pool.Query(ctx, query, politicianID, limit)
	if err != nil {
		return nil, fmt.Errorf("get sentiment: %w", err)
	}
	defer rows.Close()

	var snapshots []models.SentimentSnapshot
	for rows.Next() {
		var s models.SentimentSnapshot
		if err := rows.Scan(&s.ID, &s.PoliticianID, &s.Date, &s.Platform, &s.Score, &s.SampleSize, &s.PositivePct, &s.NegativePct, &s.NeutralPct, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sentiment: %w", err)
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

func (r *SentimentRepo) GetNationalSentiment(ctx context.Context) ([]models.SentimentSnapshot, error) {
	query := `
		SELECT DISTINCT ON (ss.politician_id)
		       ss.id, ss.politician_id, ss.date, ss.platform, ss.score, ss.sample_size,
		       ss.positive_pct, ss.negative_pct, ss.neutral_pct, ss.created_at
		FROM sentiment_snapshots ss
		WHERE ss.platform = 'overall'
		ORDER BY ss.politician_id, ss.date DESC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get national sentiment: %w", err)
	}
	defer rows.Close()

	var snapshots []models.SentimentSnapshot
	for rows.Next() {
		var s models.SentimentSnapshot
		if err := rows.Scan(&s.ID, &s.PoliticianID, &s.Date, &s.Platform, &s.Score, &s.SampleSize, &s.PositivePct, &s.NegativePct, &s.NeutralPct, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sentiment: %w", err)
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}
