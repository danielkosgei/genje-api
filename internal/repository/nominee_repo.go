package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type NomineeRepo struct {
	pool *pgxpool.Pool
}

func NewNomineeRepo(pool *pgxpool.Pool) *NomineeRepo {
	return &NomineeRepo{pool: pool}
}

// nomineeWhere builds the shared WHERE clause for list and count. Filters are bound as parameters.
func nomineeWhere(f models.NomineeFilter) (string, []any) {
	clause := " WHERE TRUE"
	args := []any{}
	add := func(cond string, value any) {
		args = append(args, value)
		clause += fmt.Sprintf(" AND "+cond, len(args))
	}
	if f.Party != "" {
		add("LOWER(party) = LOWER($%d)", f.Party)
	}
	if f.County != "" {
		add("LOWER(county) = LOWER($%d)", f.County)
	}
	if f.Heading != "" {
		add("heading = $%d", f.Heading)
	}
	if f.Query != "" {
		add("nominee_name ILIKE '%%' || $%d || '%%' ESCAPE '\\'", escapeLike(f.Query))
	}
	return clause, args
}

func (r *NomineeRepo) List(ctx context.Context, f models.NomineeFilter) ([]models.Nominee, int, error) {
	where, args := nomineeWhere(f)

	var total int
	countQuery := `SELECT COUNT(*) FROM nominee_lists` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count nominees: %w", err)
	}

	args = append(args, f.Limit, f.Offset)
	query := fmt.Sprintf(`
		SELECT id::text, source, party, heading, nominee_name, county, source_page
		FROM nominee_lists%s
		ORDER BY heading, source_page, nominee_name
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list nominees: %w", err)
	}
	defer rows.Close()

	var nominees []models.Nominee
	for rows.Next() {
		var n models.Nominee
		if err := rows.Scan(&n.ID, &n.Source, &n.Party, &n.Heading, &n.Name, &n.County, &n.SourcePage); err != nil {
			return nil, 0, fmt.Errorf("scan nominee: %w", err)
		}
		nominees = append(nominees, n)
	}
	return nominees, total, rows.Err()
}

func (r *NomineeRepo) ListHeadings(ctx context.Context) ([]models.NomineeHeading, error) {
	query := `
		SELECT heading, COUNT(*)
		FROM nominee_lists
		GROUP BY heading
		ORDER BY COUNT(*) DESC, heading`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list nominee headings: %w", err)
	}
	defer rows.Close()

	var headings []models.NomineeHeading
	for rows.Next() {
		var h models.NomineeHeading
		if err := rows.Scan(&h.Heading, &h.Count); err != nil {
			return nil, fmt.Errorf("scan nominee heading: %w", err)
		}
		headings = append(headings, h)
	}
	return headings, rows.Err()
}
