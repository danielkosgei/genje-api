package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"jalada/internal/models"
)

type PoliticianRepo struct {
	pool *pgxpool.Pool
}

func NewPoliticianRepo(pool *pgxpool.Pool) *PoliticianRepo {
	return &PoliticianRepo{pool: pool}
}

func (r *PoliticianRepo) List(ctx context.Context, f models.PoliticianFilter) ([]models.PoliticianSummary, int, error) {
	if f.Limit <= 0 {
		f.Limit = 20
	}

	countQuery := `SELECT COUNT(*) FROM politicians p WHERE 1=1`
	dataQuery := `
		SELECT p.id, p.slug, p.first_name, p.last_name, p.status, p.photo_url,
		       (SELECT pp.name FROM party_memberships pm
		        JOIN political_parties pp ON pp.id = pm.party_id
		        WHERE pm.politician_id = p.id AND pm.left_date IS NULL
		        ORDER BY pm.joined_date DESC LIMIT 1) as current_party
		FROM politicians p WHERE 1=1`

	var args []interface{}
	argIdx := 1
	where := ""

	if f.Query != "" {
		where += fmt.Sprintf(` AND (p.first_name || ' ' || p.last_name || ' ' || COALESCE(p.other_names, '')) ILIKE $%d`, argIdx)
		args = append(args, "%"+f.Query+"%")
		argIdx++
	}
	if f.PartyID != nil {
		where += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM party_memberships pm WHERE pm.politician_id = p.id AND pm.party_id = $%d AND pm.left_date IS NULL)`, argIdx)
		args = append(args, *f.PartyID)
		argIdx++
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count politicians: %w", err)
	}

	dataQuery += where + fmt.Sprintf(` ORDER BY p.last_name, p.first_name LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, f.Limit, f.Offset)

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list politicians: %w", err)
	}
	defer rows.Close()

	var politicians []models.PoliticianSummary
	for rows.Next() {
		var p models.PoliticianSummary
		if err := rows.Scan(&p.ID, &p.Slug, &p.FirstName, &p.LastName, &p.Status, &p.PhotoURL, &p.Party); err != nil {
			return nil, 0, fmt.Errorf("scan politician: %w", err)
		}
		politicians = append(politicians, p)
	}

	return politicians, total, nil
}

func (r *PoliticianRepo) GetBySlug(ctx context.Context, slug string) (*models.Politician, error) {
	query := `
		SELECT id, slug, first_name, last_name, other_names, date_of_birth, date_of_death,
		       gender, status, bio, photo_url, education, career_history, created_at, updated_at
		FROM politicians WHERE slug = $1`

	var p models.Politician
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.Slug, &p.FirstName, &p.LastName, &p.OtherNames,
		&p.DateOfBirth, &p.DateOfDeath, &p.Gender, &p.Status, &p.Bio, &p.PhotoURL,
		&p.Education, &p.CareerHistory, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get politician by slug: %w", err)
	}
	return &p, nil
}

func (r *PoliticianRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Politician, error) {
	query := `
		SELECT id, slug, first_name, last_name, other_names, date_of_birth,
		       gender, bio, photo_url, education, career_history, created_at, updated_at
		FROM politicians WHERE id = $1`

	var p models.Politician
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Slug, &p.FirstName, &p.LastName, &p.OtherNames,
		&p.DateOfBirth, &p.Gender, &p.Bio, &p.PhotoURL,
		&p.Education, &p.CareerHistory, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get politician by id: %w", err)
	}
	return &p, nil
}

func (r *PoliticianRepo) GetPartyHistory(ctx context.Context, politicianID uuid.UUID) ([]models.PartyMembership, error) {
	query := `
		SELECT pm.id, pm.politician_id, pm.party_id, pp.name, pm.joined_date, pm.left_date, pm.role, pm.created_at
		FROM party_memberships pm
		JOIN political_parties pp ON pp.id = pm.party_id
		WHERE pm.politician_id = $1
		ORDER BY pm.joined_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get party history: %w", err)
	}
	defer rows.Close()

	var memberships []models.PartyMembership
	for rows.Next() {
		var m models.PartyMembership
		if err := rows.Scan(&m.ID, &m.PoliticianID, &m.PartyID, &m.PartyName, &m.JoinedDate, &m.LeftDate, &m.Role, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan party membership: %w", err)
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}

func (r *PoliticianRepo) GetCandidacies(ctx context.Context, politicianID uuid.UUID) ([]models.CandidacyDetail, error) {
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
		WHERE c.politician_id = $1
		ORDER BY e.election_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get candidacies: %w", err)
	}
	defer rows.Close()

	var candidacies []models.CandidacyDetail
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
		candidacies = append(candidacies, cd)
	}
	return candidacies, nil
}

func (r *PoliticianRepo) GetIntegrityFlags(ctx context.Context, politicianID uuid.UUID) ([]models.IntegrityFlag, error) {
	query := `
		SELECT id, politician_id, flag_type, description, status, source_url, source_id, flagged_at, created_at, updated_at
		FROM integrity_flags
		WHERE politician_id = $1
		ORDER BY flagged_at DESC`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get integrity flags: %w", err)
	}
	defer rows.Close()

	var flags []models.IntegrityFlag
	for rows.Next() {
		var f models.IntegrityFlag
		if err := rows.Scan(&f.ID, &f.PoliticianID, &f.FlagType, &f.Description, &f.Status, &f.SourceURL, &f.SourceID, &f.FlaggedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan integrity flag: %w", err)
		}
		flags = append(flags, f)
	}
	return flags, nil
}

func (r *PoliticianRepo) GetCourtCases(ctx context.Context, politicianID uuid.UUID) ([]models.CourtCase, error) {
	query := `
		SELECT id, politician_id, case_number, court_name, case_type, title, description,
		       filing_date, status, outcome, source_url, source_id, created_at, updated_at
		FROM court_cases
		WHERE politician_id = $1
		ORDER BY filing_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get court cases: %w", err)
	}
	defer rows.Close()

	var cases []models.CourtCase
	for rows.Next() {
		var c models.CourtCase
		if err := rows.Scan(
			&c.ID, &c.PoliticianID, &c.CaseNumber, &c.CourtName, &c.CaseType, &c.Title, &c.Description,
			&c.FilingDate, &c.Status, &c.Outcome, &c.SourceURL, &c.SourceID, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan court case: %w", err)
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func (r *PoliticianRepo) GetPromises(ctx context.Context, politicianID uuid.UUID) ([]models.Promise, error) {
	query := `
		SELECT id, politician_id, description, sector, made_date, deadline, status,
		       evidence, source_url, source_id, created_at, updated_at
		FROM promises
		WHERE politician_id = $1
		ORDER BY made_date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get promises: %w", err)
	}
	defer rows.Close()

	var promises []models.Promise
	for rows.Next() {
		var p models.Promise
		if err := rows.Scan(
			&p.ID, &p.PoliticianID, &p.Description, &p.Sector, &p.MadeDate, &p.Deadline,
			&p.Status, &p.Evidence, &p.SourceURL, &p.SourceID, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan promise: %w", err)
		}
		promises = append(promises, p)
	}
	return promises, nil
}

func (r *PoliticianRepo) GetPromiseStats(ctx context.Context, politicianID uuid.UUID) (*models.PromiseStats, error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'fulfilled'),
			COUNT(*) FILTER (WHERE status = 'broken'),
			COUNT(*) FILTER (WHERE status = 'in_progress'),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'partially_fulfilled')
		FROM promises WHERE politician_id = $1`

	var s models.PromiseStats
	err := r.pool.QueryRow(ctx, query, politicianID).Scan(
		&s.Total, &s.Fulfilled, &s.Broken, &s.InProgress, &s.Pending, &s.PartiallyFulfilled,
	)
	if err != nil {
		return nil, fmt.Errorf("get promise stats: %w", err)
	}
	if s.Total > 0 {
		s.FulfillmentRate = float64(s.Fulfilled) / float64(s.Total) * 100
	}
	return &s, nil
}

func (r *PoliticianRepo) GetAchievements(ctx context.Context, politicianID uuid.UUID) ([]models.Achievement, error) {
	query := `
		SELECT id, politician_id, title, description, category, date, source_url, source_id, created_at
		FROM achievements WHERE politician_id = $1
		ORDER BY date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []models.Achievement
	for rows.Next() {
		var a models.Achievement
		if err := rows.Scan(&a.ID, &a.PoliticianID, &a.Title, &a.Description, &a.Category, &a.Date, &a.SourceURL, &a.SourceID, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	return achievements, nil
}

func (r *PoliticianRepo) GetControversies(ctx context.Context, politicianID uuid.UUID) ([]models.Controversy, error) {
	query := `
		SELECT id, politician_id, title, description, category, date, severity, source_url, source_id, created_at
		FROM controversies WHERE politician_id = $1
		ORDER BY date DESC NULLS LAST`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get controversies: %w", err)
	}
	defer rows.Close()

	var controversies []models.Controversy
	for rows.Next() {
		var c models.Controversy
		if err := rows.Scan(&c.ID, &c.PoliticianID, &c.Title, &c.Description, &c.Category, &c.Date, &c.Severity, &c.SourceURL, &c.SourceID, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan controversy: %w", err)
		}
		controversies = append(controversies, c)
	}
	return controversies, nil
}

func (r *PoliticianRepo) GetAffiliations(ctx context.Context, politicianID uuid.UUID) ([]models.AffiliationDetail, error) {
	query := `
		SELECT a.id, a.politician_id, a.related_politician_id, a.relationship_type, a.description,
		       a.source_url, a.source_id, a.created_at,
		       p.first_name || ' ' || p.last_name, p.slug, p.photo_url
		FROM affiliations a
		JOIN politicians p ON p.id = a.related_politician_id
		WHERE a.politician_id = $1
		ORDER BY a.relationship_type, a.created_at DESC`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get affiliations: %w", err)
	}
	defer rows.Close()

	var affiliations []models.AffiliationDetail
	for rows.Next() {
		var ad models.AffiliationDetail
		if err := rows.Scan(
			&ad.ID, &ad.PoliticianID, &ad.RelatedPoliticianID, &ad.RelationshipType, &ad.Description,
			&ad.SourceURL, &ad.SourceID, &ad.CreatedAt,
			&ad.RelatedPoliticianName, &ad.RelatedPoliticianSlug, &ad.RelatedPhotoURL,
		); err != nil {
			return nil, fmt.Errorf("scan affiliation: %w", err)
		}
		affiliations = append(affiliations, ad)
	}
	return affiliations, nil
}

func (r *PoliticianRepo) GetAssetDeclarations(ctx context.Context, politicianID uuid.UUID) ([]models.AssetDeclaration, error) {
	query := `
		SELECT id, politician_id, declaration_year, total_assets, total_liabilities, details,
		       source_url, source_id, created_at
		FROM asset_declarations WHERE politician_id = $1
		ORDER BY declaration_year DESC`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get asset declarations: %w", err)
	}
	defer rows.Close()

	var declarations []models.AssetDeclaration
	for rows.Next() {
		var d models.AssetDeclaration
		if err := rows.Scan(&d.ID, &d.PoliticianID, &d.DeclarationYear, &d.TotalAssets, &d.TotalLiabilities, &d.Details, &d.SourceURL, &d.SourceID, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan asset declaration: %w", err)
		}
		declarations = append(declarations, d)
	}
	return declarations, nil
}

func (r *PoliticianRepo) GetVotingRecords(ctx context.Context, politicianID uuid.UUID) ([]models.VotingRecord, error) {
	query := `
		SELECT id, politician_id, bill_name, bill_number, vote, vote_date, session, source_url, created_at
		FROM voting_records WHERE politician_id = $1
		ORDER BY vote_date DESC`

	rows, err := r.pool.Query(ctx, query, politicianID)
	if err != nil {
		return nil, fmt.Errorf("get voting records: %w", err)
	}
	defer rows.Close()

	var records []models.VotingRecord
	for rows.Next() {
		var v models.VotingRecord
		if err := rows.Scan(&v.ID, &v.PoliticianID, &v.BillName, &v.BillNumber, &v.Vote, &v.VoteDate, &v.Session, &v.SourceURL, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan voting record: %w", err)
		}
		records = append(records, v)
	}
	return records, nil
}

func (r *PoliticianRepo) GetAttendanceStats(ctx context.Context, politicianID uuid.UUID) (*models.AttendanceStats, error) {
	query := `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE present = true),
			COUNT(*) FILTER (WHERE present = false)
		FROM parliamentary_attendance WHERE politician_id = $1`

	var s models.AttendanceStats
	err := r.pool.QueryRow(ctx, query, politicianID).Scan(&s.TotalSessions, &s.Present, &s.Absent)
	if err != nil {
		return nil, fmt.Errorf("get attendance stats: %w", err)
	}
	if s.TotalSessions > 0 {
		s.AttendanceRate = float64(s.Present) / float64(s.TotalSessions) * 100
	}
	return &s, nil
}
