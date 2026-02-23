package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepo struct {
	pool *pgxpool.Pool
}

func NewAnalyticsRepo(pool *pgxpool.Pool) *AnalyticsRepo {
	return &AnalyticsRepo{pool: pool}
}

type PromiseAnalytics struct {
	TotalPoliticians   int     `json:"total_politicians"`
	TotalPromises      int     `json:"total_promises"`
	FulfilledCount     int     `json:"fulfilled_count"`
	BrokenCount        int     `json:"broken_count"`
	PendingCount       int     `json:"pending_count"`
	InProgressCount    int     `json:"in_progress_count"`
	OverallFulfillment float64 `json:"overall_fulfillment_rate"`
}

type IntegrityAnalytics struct {
	TotalPoliticians int `json:"total_politicians"`
	WithCourtCases   int `json:"with_court_cases"`
	WithFlags        int `json:"with_integrity_flags"`
	Chapter6Issues   int `json:"chapter6_issues"`
	PendingCases     int `json:"pending_cases"`
}

type AttendanceAnalytics struct {
	TotalPoliticians    int     `json:"total_politicians"`
	AverageAttendance   float64 `json:"average_attendance_rate"`
	HighestAttendance   float64 `json:"highest_attendance_rate"`
	LowestAttendance    float64 `json:"lowest_attendance_rate"`
}

type TrendingItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Score int    `json:"score"`
	Type  string `json:"type"`
}

func (r *AnalyticsRepo) GetPromiseAnalytics(ctx context.Context) (*PromiseAnalytics, error) {
	query := `
		SELECT
			(SELECT COUNT(DISTINCT politician_id) FROM promises),
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'fulfilled'),
			COUNT(*) FILTER (WHERE status = 'broken'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'in_progress')
		FROM promises`

	var pa PromiseAnalytics
	err := r.pool.QueryRow(ctx, query).Scan(
		&pa.TotalPoliticians, &pa.TotalPromises,
		&pa.FulfilledCount, &pa.BrokenCount, &pa.PendingCount, &pa.InProgressCount,
	)
	if err != nil {
		return nil, fmt.Errorf("get promise analytics: %w", err)
	}
	if pa.TotalPromises > 0 {
		pa.OverallFulfillment = float64(pa.FulfilledCount) / float64(pa.TotalPromises) * 100
	}
	return &pa, nil
}

func (r *AnalyticsRepo) GetIntegrityAnalytics(ctx context.Context) (*IntegrityAnalytics, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM politicians),
			(SELECT COUNT(DISTINCT politician_id) FROM court_cases),
			(SELECT COUNT(DISTINCT politician_id) FROM integrity_flags WHERE status = 'active'),
			(SELECT COUNT(*) FROM integrity_flags WHERE flag_type = 'chapter6' AND status = 'active'),
			(SELECT COUNT(*) FROM court_cases WHERE status IN ('pending', 'ongoing'))
		`

	var ia IntegrityAnalytics
	err := r.pool.QueryRow(ctx, query).Scan(
		&ia.TotalPoliticians, &ia.WithCourtCases, &ia.WithFlags,
		&ia.Chapter6Issues, &ia.PendingCases,
	)
	if err != nil {
		return nil, fmt.Errorf("get integrity analytics: %w", err)
	}
	return &ia, nil
}

func (r *AnalyticsRepo) GetAttendanceAnalytics(ctx context.Context) (*AttendanceAnalytics, error) {
	query := `
		WITH rates AS (
			SELECT politician_id,
			       COUNT(*) FILTER (WHERE present = true)::float / NULLIF(COUNT(*), 0) * 100 as rate
			FROM parliamentary_attendance
			GROUP BY politician_id
		)
		SELECT
			COUNT(*),
			COALESCE(AVG(rate), 0),
			COALESCE(MAX(rate), 0),
			COALESCE(MIN(rate), 0)
		FROM rates`

	var aa AttendanceAnalytics
	err := r.pool.QueryRow(ctx, query).Scan(
		&aa.TotalPoliticians, &aa.AverageAttendance, &aa.HighestAttendance, &aa.LowestAttendance,
	)
	if err != nil {
		return nil, fmt.Errorf("get attendance analytics: %w", err)
	}
	return &aa, nil
}

func (r *AnalyticsRepo) GetTrending(ctx context.Context, limit int) ([]TrendingItem, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT p.id::text, p.first_name || ' ' || p.last_name, p.slug, COUNT(apm.article_id) as mentions
		FROM politicians p
		JOIN article_politician_mentions apm ON apm.politician_id = p.id
		JOIN news_articles na ON na.id = apm.article_id AND na.published_at > NOW() - INTERVAL '7 days'
		GROUP BY p.id, p.first_name, p.last_name, p.slug
		ORDER BY mentions DESC
		LIMIT $1`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get trending: %w", err)
	}
	defer rows.Close()

	var items []TrendingItem
	for rows.Next() {
		var t TrendingItem
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Score); err != nil {
			return nil, fmt.Errorf("scan trending: %w", err)
		}
		t.Type = "politician"
		items = append(items, t)
	}
	return items, nil
}
