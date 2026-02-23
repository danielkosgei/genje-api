package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(pool *pgxpool.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

func (r *EventRepo) ListEvents(ctx context.Context, limit, offset int) ([]models.Event, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM events`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	query := `
		SELECT id, title, description, event_type, location, latitude, longitude,
		       start_time, end_time, source_url, source_id, created_at, updated_at
		FROM events
		ORDER BY start_time DESC NULLS LAST
		LIMIT $1 OFFSET $2`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EventType, &e.Location,
			&e.Latitude, &e.Longitude, &e.StartTime, &e.EndTime,
			&e.SourceURL, &e.SourceID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, nil
}

func (r *EventRepo) GetTimelineEvents(ctx context.Context) ([]models.TimelineEvent, error) {
	query := `
		SELECT et.id, et.election_id, et.title, et.description, et.milestone_type,
		       et.date, et.status, et.created_at, et.updated_at
		FROM election_timeline et
		JOIN elections e ON e.id = et.election_id
		WHERE e.status != 'completed'
		ORDER BY et.date`

	rows, err := r.pool.Query(ctx, query)
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

func (r *EventRepo) GetEventsByPolitician(ctx context.Context, politicianID interface{}, limit, offset int) ([]models.Event, int, error) {
	if limit <= 0 {
		limit = 20
	}

	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM events e JOIN event_participants ep ON ep.event_id = e.id WHERE ep.politician_id = $1`,
		politicianID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count politician events: %w", err)
	}

	query := `
		SELECT e.id, e.title, e.description, e.event_type, e.location, e.latitude, e.longitude,
		       e.start_time, e.end_time, e.source_url, e.source_id, e.created_at, e.updated_at
		FROM events e
		JOIN event_participants ep ON ep.event_id = e.id
		WHERE ep.politician_id = $1
		ORDER BY e.start_time DESC NULLS LAST
		LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, politicianID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list politician events: %w", err)
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EventType, &e.Location,
			&e.Latitude, &e.Longitude, &e.StartTime, &e.EndTime,
			&e.SourceURL, &e.SourceID, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, total, nil
}
