package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"jalada/internal/models"
)

// getLinks returns a politician's cited links: profile pages and websites, each with its publisher and check date.
func (r *PoliticianRepo) getLinks(ctx context.Context, politicianID uuid.UUID) ([]models.PoliticianLink, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT pl.kind, pl.title, pl.url, s.name, pl.publisher_type,
		       to_char(pl.checked_on, 'YYYY-MM-DD'), pl.check_note
		FROM politician_links pl
		JOIN sources s ON s.id = pl.source_id
		WHERE pl.politician_id = $1
		ORDER BY pl.kind, pl.title`, politicianID)
	if err != nil {
		return nil, fmt.Errorf("politician links: %w", err)
	}
	defer rows.Close()

	links := []models.PoliticianLink{}
	for rows.Next() {
		var l models.PoliticianLink
		if err := rows.Scan(&l.Kind, &l.Title, &l.URL, &l.Publisher, &l.PublisherType, &l.CheckedOn, &l.CheckNote); err != nil {
			return nil, fmt.Errorf("scan politician link: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}
