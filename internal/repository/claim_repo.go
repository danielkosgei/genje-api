package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"jalada/internal/models"
)

// GetClaims returns one politician's rated claims, most recently rated first.
func (r *PoliticianRepo) GetClaims(ctx context.Context, politicianID uuid.UUID) ([]models.Claim, error) {
	query := `
		SELECT id::text, claim, rating,
		       to_char(rated_on, 'YYYY-MM-DD'), source_title, source_url
		FROM claims
		WHERE politician_id = $1
		ORDER BY rated_on DESC NULLS LAST, id`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get claims: %w", err)
	}
	defer rows.Close()

	var claims []models.Claim
	for rows.Next() {
		var c models.Claim
		if err := rows.Scan(&c.ID, &c.Claim, &c.Rating, &c.RatedOn, &c.SourceTitle, &c.SourceURL); err != nil {
			return nil, fmt.Errorf("scan claim: %w", err)
		}
		claims = append(claims, c)
	}
	return claims, rows.Err()
}
