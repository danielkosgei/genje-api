package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"jalada/internal/models"
)

// PoliticianProfile reads the figures a dossier shows beside the record. Anything not yet sourced comes back null.
// Current post and years in office count candidacies with status 'elected'; ingestion must use that status for winners.
func (r *PoliticianRepo) GetProfile(ctx context.Context, id uuid.UUID, slug string) (models.PoliticianProfile, error) {
	profile := models.PoliticianProfile{Slug: slug, ClaimCounts: models.ClaimCounts{}, PolicyPositions: []models.ProfilePolicyPosition{}}

	if err := r.pool.QueryRow(ctx, `
		SELECT CASE WHEN date_of_birth IS NULL THEN NULL
		            ELSE date_part('year', age(current_date, date_of_birth))::int END
		FROM politicians WHERE id = $1`, id).Scan(&profile.Age); err != nil {
		return profile, fmt.Errorf("profile age: %w", err)
	}

	var post *string
	var years *int
	err := r.pool.QueryRow(ctx, `
		SELECT
		  (SELECT el.title FROM candidacies c
		     JOIN elective_positions el ON el.id = c.position_id
		     JOIN elections e ON e.id = c.election_id
		     WHERE c.politician_id = $1 AND c.status = 'elected'
		     ORDER BY e.election_date DESC NULLS LAST LIMIT 1),
		  (SELECT date_part('year', age(current_date, MIN(e.election_date)))::int FROM candidacies c
		     JOIN elections e ON e.id = c.election_id
		     WHERE c.politician_id = $1 AND c.status = 'elected' AND e.election_date IS NOT NULL)`, id).Scan(&post, &years)
	if err != nil {
		return profile, fmt.Errorf("profile office: %w", err)
	}
	profile.CurrentPost = post
	profile.YearsInOffice = years

	rows, err := r.pool.Query(ctx, `SELECT rating, COUNT(*) FROM claims WHERE politician_id = $1 GROUP BY rating`, id)
	if err != nil {
		return profile, fmt.Errorf("profile claims: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var rating string
		var n int
		if err := rows.Scan(&rating, &n); err != nil {
			return profile, fmt.Errorf("scan claim count: %w", err)
		}
		switch rating {
		case "confirmed":
			profile.ClaimCounts.Confirmed = n
		case "unproven":
			profile.ClaimCounts.Unproven = n
		case "contested":
			profile.ClaimCounts.Contested = n
		}
	}
	if err := rows.Err(); err != nil {
		return profile, err
	}

	posRows, err := r.pool.Query(ctx, `
		SELECT pp.sector, pp.title, pp.description, pp.source_url
		FROM policy_positions pp
		JOIN manifestos m ON m.id = pp.manifesto_id
		WHERE m.politician_id = $1
		ORDER BY pp.sector, pp.title`, id)
	if err != nil {
		return profile, fmt.Errorf("profile positions: %w", err)
	}
	defer posRows.Close()
	for posRows.Next() {
		var p models.ProfilePolicyPosition
		if err := posRows.Scan(&p.Sector, &p.Title, &p.Description, &p.SourceURL); err != nil {
			return profile, fmt.Errorf("scan policy position: %w", err)
		}
		profile.PolicyPositions = append(profile.PolicyPositions, p)
	}
	if err := posRows.Err(); err != nil {
		return profile, err
	}
	links, err := r.getLinks(ctx, id)
	if err != nil {
		return profile, err
	}
	profile.Links = links
	return profile, nil
}

// PartyProfile reads the figures for a party page. Members are current memberships; aspirants are distinct
// candidates; counties fielding are distinct counties where the party has a candidate.
func (r *PartyRepo) GetProfile(ctx context.Context, slug string) (*models.PartyProfile, error) {
	var partyID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM political_parties WHERE slug = $1`, slug).Scan(&partyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find party: %w", err)
	}

	profile := &models.PartyProfile{Slug: slug}
	if err := r.pool.QueryRow(ctx, `
		SELECT
		  (SELECT COUNT(*) FROM party_memberships WHERE party_id = $1 AND left_date IS NULL)::int,
		  (SELECT COUNT(DISTINCT politician_id) FROM candidacies WHERE party_id = $1)::int,
		  (SELECT COUNT(DISTINCT ep.county_id) FROM candidacies c
		     JOIN elective_positions ep ON ep.id = c.position_id
		     WHERE c.party_id = $1 AND ep.county_id IS NOT NULL)::int`, partyID).
		Scan(&profile.Members, &profile.Aspirants, &profile.CountiesFielding); err != nil {
		return nil, fmt.Errorf("party counts: %w", err)
	}

	err = r.pool.QueryRow(ctx, `
		SELECT p.slug, TRIM(p.first_name || ' ' || p.last_name)
		FROM political_parties pp
		JOIN politicians p ON p.id = pp.leader_id
		WHERE pp.id = $1`, partyID).Scan(&profile.LeaderSlug, &profile.LeaderName)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("party leader: %w", err)
	}
	return profile, nil
}

// Search matches names across politicians, parties, counties, coalitions and elections.
// Each group is capped at limit. The query is matched as a plain substring, never as a pattern.
func (r *SearchRepo) Search(ctx context.Context, query string, limit int) (models.SearchResults, error) {
	pattern := "%" + escapeLike(strings.TrimSpace(query)) + "%"
	out := models.SearchResults{
		Politicians: []models.SearchHit{},
		Parties:     []models.SearchHit{},
		Counties:    []models.SearchHit{},
		Coalitions:  []models.SearchHit{},
		Elections:   []models.SearchHit{},
	}

	if err := r.collect(ctx, &out.Politicians, "politician", `
		SELECT p.slug, TRIM(p.first_name || ' ' || p.last_name),
		       (SELECT pt.name FROM party_memberships pm
		          JOIN political_parties pt ON pt.id = pm.party_id
		          WHERE pm.politician_id = p.id AND pm.left_date IS NULL
		          ORDER BY pm.joined_date DESC NULLS LAST LIMIT 1)
		FROM politicians p
		WHERE (p.first_name || ' ' || p.last_name) ILIKE $1 ESCAPE '\' OR p.other_names ILIKE $1 ESCAPE '\'
		ORDER BY p.last_name, p.first_name LIMIT $2`, pattern, limit); err != nil {
		return out, err
	}
	if err := r.collect(ctx, &out.Parties, "party", `
		SELECT slug, name, abbreviation FROM political_parties
		WHERE name ILIKE $1 ESCAPE '\' OR abbreviation ILIKE $1 ESCAPE '\'
		ORDER BY name LIMIT $2`, pattern, limit); err != nil {
		return out, err
	}
	if err := r.collect(ctx, &out.Counties, "county", `
		SELECT slug, name, code FROM counties
		WHERE name ILIKE $1 ESCAPE '\' ORDER BY name LIMIT $2`, pattern, limit); err != nil {
		return out, err
	}
	if err := r.collect(ctx, &out.Coalitions, "coalition", `
		SELECT slug, name, NULL::text FROM coalitions
		WHERE name ILIKE $1 ESCAPE '\' ORDER BY name LIMIT $2`, pattern, limit); err != nil {
		return out, err
	}
	if err := r.collect(ctx, &out.Elections, "election", `
		SELECT id::text, name, NULL::text FROM elections
		WHERE name ILIKE $1 ESCAPE '\' ORDER BY election_date DESC NULLS LAST LIMIT $2`, pattern, limit); err != nil {
		return out, err
	}
	return out, nil
}

func (r *SearchRepo) collect(ctx context.Context, into *[]models.SearchHit, kind, query, pattern string, limit int) error {
	rows, err := r.pool.Query(ctx, query, pattern, limit)
	if err != nil {
		return fmt.Errorf("search %s: %w", kind, err)
	}
	defer rows.Close()
	for rows.Next() {
		var hit models.SearchHit
		hit.Kind = kind
		if err := rows.Scan(&hit.Slug, &hit.Name, &hit.Subtitle); err != nil {
			return fmt.Errorf("scan search %s: %w", kind, err)
		}
		*into = append(*into, hit)
	}
	return rows.Err()
}

// CountyRace reads the governor candidates for one county. ok is false when no county has that code.
func (r *RaceRepo) CountyRace(ctx context.Context, countyCode string) (models.CountyRace, bool, error) {
	race := models.CountyRace{CountyCode: countyCode, Candidates: []models.RaceCandidate{}}
	var countyID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM counties WHERE code = $1`, countyCode).Scan(&countyID)
	if errors.Is(err, pgx.ErrNoRows) {
		return race, false, nil
	}
	if err != nil {
		return race, false, fmt.Errorf("find county: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.slug, TRIM(p.first_name || ' ' || p.last_name), pt.abbreviation, c.status
		FROM candidacies c
		JOIN elective_positions ep ON ep.id = c.position_id
		JOIN politicians p ON p.id = c.politician_id
		LEFT JOIN political_parties pt ON pt.id = c.party_id
		WHERE ep.county_id = $1 AND ep.level = 'governor'
		ORDER BY p.last_name, p.first_name`, countyID)
	if err != nil {
		return race, true, fmt.Errorf("county race: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cand models.RaceCandidate
		if err := rows.Scan(&cand.PoliticianSlug, &cand.Name, &cand.Party, &cand.Status); err != nil {
			return race, true, fmt.Errorf("scan candidate: %w", err)
		}
		race.Candidates = append(race.Candidates, cand)
	}
	return race, true, rows.Err()
}

// PollSummary averages the latest poll from each pollster for one race, published within windowDays.
func (r *PollRepo) Summary(ctx context.Context, race string, windowDays int) (models.PollSummary, error) {
	summary := models.PollSummary{Race: race, WindowDays: windowDays, Candidates: []models.PollCandidate{}}
	err := r.pool.QueryRow(ctx, `
		WITH latest AS (
		  SELECT DISTINCT ON (pollster_id) id
		  FROM polls
		  WHERE race = $1 AND published_on >= current_date - $2::int
		  ORDER BY pollster_id, published_on DESC, id
		)
		SELECT (SELECT COUNT(*) FROM latest)::int`, race, windowDays).Scan(&summary.PollCount)
	if err != nil {
		return summary, fmt.Errorf("poll count: %w", err)
	}
	if summary.PollCount == 0 {
		return summary, nil
	}

	rows, err := r.pool.Query(ctx, `
		WITH latest AS (
		  SELECT DISTINCT ON (pollster_id) id
		  FROM polls
		  WHERE race = $1 AND published_on >= current_date - $2::int
		  ORDER BY pollster_id, published_on DESC, id
		)
		SELECT pr.candidate_name, MAX(pr.party), AVG(pr.share)::float8
		FROM poll_results pr
		JOIN latest l ON l.id = pr.poll_id
		GROUP BY pr.candidate_name
		ORDER BY AVG(pr.share) DESC, pr.candidate_name`, race, windowDays)
	if err != nil {
		return summary, fmt.Errorf("poll averages: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c models.PollCandidate
		if err := rows.Scan(&c.CandidateName, &c.Party, &c.AverageShare); err != nil {
			return summary, fmt.Errorf("scan poll average: %w", err)
		}
		summary.Candidates = append(summary.Candidates, c)
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}
	if len(summary.Candidates) >= 1 {
		leader := summary.Candidates[0]
		summary.Leader = &leader
	}
	if len(summary.Candidates) >= 2 {
		margin := summary.Candidates[0].AverageShare - summary.Candidates[1].AverageShare
		summary.Margin = &margin
	}
	return summary, nil
}
